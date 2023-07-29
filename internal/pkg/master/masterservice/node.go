package masterservice

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/wyattjychen/hades/internal/pkg/etcdconn"
	"github.com/wyattjychen/hades/internal/pkg/logger"
	"github.com/wyattjychen/hades/internal/pkg/model"
)

type NodeWatcherService struct {
	client   *etcdconn.Client
	nodeList map[string]model.Node
	lock     sync.Mutex
}

var DefaultNodeWatcher *NodeWatcherService

func NewNodeWatcherService() *NodeWatcherService {
	return &NodeWatcherService{
		client:   etcdconn.GetEtcdClient(),
		nodeList: make(map[string]model.Node),
	}
}

func (n *NodeWatcherService) Watch() error {
	resp, err := n.client.Get(context.Background(), etcdconn.EtcdNodeKeyPrefix, clientv3.WithPrefix())
	if err != nil {
		return err
	}
	_ = n.getNodes(resp)

	go n.watcher()
	return nil
}

func (n *NodeWatcherService) delNodeList(key string) {
	n.lock.Lock()
	defer n.lock.Unlock()
	delete(n.nodeList, key)
	logger.GetLogger().Debug(fmt.Sprintf("delelte node[%s]", key))
}

func (n *NodeWatcherService) watcher() {
	rch := n.client.Watch(context.Background(), etcdconn.EtcdNodeKeyPrefix, clientv3.WithPrefix())
	for wresp := range rch {
		for _, ev := range wresp.Events {
			switch ev.Type {
			case mvccpb.PUT:
				n.setNodeList(n.GetUUID(string(ev.Kv.Key)), string(ev.Kv.Value))
			case mvccpb.DELETE:
				uuid := n.GetUUID(string(ev.Kv.Key))
				n.delNodeList(uuid)
				logger.GetLogger().Warn(fmt.Sprintf("crony node[%s] DELETE event detected", uuid))
				node := &model.Node{UUID: uuid}
				err := node.FindByUUID()
				if err != nil {
					logger.GetLogger().Error(fmt.Sprintf("crony node[%s] find by uuid  error:%s", uuid, err.Error()))
					return
				}

				_, fail, err := n.FailOver(uuid)
				if err != nil {
					logger.GetLogger().Error(fmt.Sprintf("crony node[%s] fail over error:%s", uuid, err.Error()))
					return
				}
				// if the failover is all successful, delete the node in the database
				if fail.Count() == 0 {
					err = node.Delete()
					if err != nil {
						logger.GetLogger().Error(fmt.Sprintf("crony node[%s] delete by uuid  error:%s", uuid, err.Error()))
					}
				}

			}
		}
	}
}

type Result []int

func (r Result) Count() (count int) {
	for _, v := range r {
		if v != 0 {
			count++
		}
	}
	return
}

func (n *NodeWatcherService) FailOver(nodeUUID string) (success Result, fail Result, err error) {
	jobs, err := n.GetJobs(nodeUUID)
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("node[%s] fail over get jobs error:%s", nodeUUID, err.Error()))
		return
	}
	if len(jobs) == 0 {
		return
	}
	for _, job := range jobs {

		oldUUID := job.RunOn
		autoUUID := DefaultJobService.AutoAllocateNode()
		if autoUUID == "" {
			logger.GetLogger().Warn(fmt.Sprintf("node[%s] job[%d] fail over auto allocate node error", nodeUUID, job.ID))
			fail = append(fail, job.ID)
			continue
		}
		err = n.assignJob(autoUUID, &job)
		if err != nil {
			logger.GetLogger().Warn(fmt.Sprintf("node[%s] job[%d] fail over assign job error", nodeUUID, job.ID))
			fail = append(fail, job.ID)
			continue
		}
		//Delete the key value if the transfer is successful
		_, err = etcdconn.Delete(fmt.Sprintf(etcdconn.EtcdJobKey, oldUUID, job.ID))
		if err != nil {
			logger.GetLogger().Error(fmt.Sprintf("node[%s] job[%d] fail over etcd delete job error:%s", nodeUUID, job.ID, err.Error()))
			fail = append(fail, job.ID)
			continue
		}
		success = append(success, job.ID)
	}
	return
}

// get all the job under a node
func (n *NodeWatcherService) GetJobs(nodeUUID string) (jobs []model.Job, err error) {
	resps, err := etcdconn.Get(fmt.Sprintf(etcdconn.EtcdJobKeyPrefix, nodeUUID), clientv3.WithPrefix())
	if err != nil {
		return
	}
	count := len(resps.Kvs)
	if count == 0 {
		return
	}
	for _, j := range resps.Kvs {
		var job model.Job
		if err := json.Unmarshal(j.Value, &job); err != nil {
			logger.GetLogger().Warn(fmt.Sprintf("job[%s] umarshal err: %s", string(j.Key), err.Error()))
			continue
		}
		jobs = append(jobs, job)
	}
	return
}

func (n *NodeWatcherService) getNodes(resp *clientv3.GetResponse) []string {
	nodes := make([]string, 0)
	if resp == nil || resp.Kvs == nil {
		return nodes
	}
	for i := range resp.Kvs {
		if v := resp.Kvs[i].Value; v != nil {
			n.setNodeList(n.GetUUID(string(resp.Kvs[i].Key)), string(resp.Kvs[i].Value))
			nodes = append(nodes, string(v))
		}
	}
	return nodes
}

func (n *NodeWatcherService) GetUUID(key string) string {
	// /hades/node/<node_uuid>
	index := strings.LastIndex(key, "/")
	if index == -1 {
		return ""
	}
	return key[index+1:]
}

func (n *NodeWatcherService) List2Array() []string {
	n.lock.Lock()
	defer n.lock.Unlock()
	nodes := make([]string, 0)

	for k, _ := range n.nodeList {
		nodes = append(nodes, k)
	}
	return nodes
}

func (n *NodeWatcherService) GetJobCount(nodeUUID string) (int, error) {
	resps, err := etcdconn.Get(fmt.Sprintf(etcdconn.EtcdJobKeyPrefix, nodeUUID), clientv3.WithPrefix(), clientv3.WithCountOnly())
	if err != nil {
		return 0, err
	}
	return int(resps.Count), nil
}

func (n *NodeWatcherService) setNodeList(key, value string) {
	var node model.Node
	err := json.Unmarshal([]byte(value), &node)
	if err != nil {
		logger.GetLogger().Warn(fmt.Sprintf("discover node[%s] json error:%s", key, err.Error()))
		return
	}
	n.lock.Lock()
	n.nodeList[key] = node
	n.lock.Unlock()
	logger.GetLogger().Debug(fmt.Sprintf("discover node node[%s] with pid[%s]", key, value))
	//Wait for the node to be fully started and assign the node
	time.Sleep(5 * time.Second)
	//find unassigned job
	jobs, err := DefaultJobService.GetNotAssignedJob()
	if err != nil {
		logger.GetLogger().Warn(fmt.Sprintf("discover node[%s],pid[%s] and get not assigned job err:%s", key, value, err.Error()))
		return
	}
	for _, job := range jobs {

		err = job.Unmarshal()
		if err != nil {
			logger.GetLogger().Warn(fmt.Sprintf("assign unassigned job[%d] json unmarshal error:%s", job.ID, err.Error()))
			continue
		}
		oldUUID := job.RunOn
		nodeUUID := DefaultJobService.AutoAllocateNode()
		if nodeUUID == "" {
			//If automatic allocation fails, it will be directly assigned to the new node.
			nodeUUID = key
		}
		err = n.assignJob(nodeUUID, &job)
		if err != nil {
			logger.GetLogger().Warn(fmt.Sprintf("assign unassigned job[%d]  error:%s", job.ID, err.Error()))
			continue
		}
		//Delete the key value if the transfer is successful
		_, err = etcdconn.Delete(fmt.Sprintf(etcdconn.EtcdJobKey, oldUUID, job.ID))
		if err != nil {
			logger.GetLogger().Error(fmt.Sprintf("node[%s] job[%d] fail over etcd delete job error:%s", nodeUUID, job.ID, err.Error()))
			continue
		}
	}
}

func (n *NodeWatcherService) assignJob(nodeUUID string, job *model.Job) (err error) {
	if nodeUUID == "" {
		return fmt.Errorf("node uuid can't be null")
	}
	node, ok := n.nodeList[nodeUUID]
	if !ok {
		return fmt.Errorf("assign unassigned job[%d] but  node[%s] not exist ", job.ID, nodeUUID)
	}
	job.InitNodeInfo(model.JobStatusAssigned, node.UUID, node.Hostname, node.IP)

	b, err := json.Marshal(job)
	if err != nil {
		return
	}
	_, err = etcdconn.Put(fmt.Sprintf(etcdconn.EtcdJobKey, nodeUUID, job.ID), string(b))
	if err != nil {
		return
	}
	err = job.Update()
	if err != nil {
		return
	}
	return
}
