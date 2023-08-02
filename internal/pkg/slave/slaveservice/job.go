package slaveservice

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/ouqiang/goutil"
	"github.com/wyattjychen/hades/internal/pkg/etcdconn"
	"github.com/wyattjychen/hades/internal/pkg/logger"
	"github.com/wyattjychen/hades/internal/pkg/model"
	"github.com/wyattjychen/hades/internal/pkg/slave/slavehandler"
	"github.com/wyattjychen/hades/internal/pkg/utils"
)

func (srv *NodeServer) modifyJob(j *slavehandler.Job) {
	oldJob, ok := srv.Jobs[j.ID]
	if !ok {
		srv.addJob(j)
		return
	}
	srv.deleteJob(oldJob.ID)
	srv.addJob(j)
	return
}

func (srv *NodeServer) deleteJob(jobId int) {
	if _, ok := srv.Jobs[jobId]; ok {
		srv.Cron.RemoveJob(srv.jobCronName(jobId))
		delete(srv.Jobs, jobId)
		return
	}
	return
}

func (srv *NodeServer) watchJobs() {
	rch := slavehandler.WatchJobs(srv.UUID)
	for wresp := range rch {
		for _, ev := range wresp.Events {
			switch {
			case ev.IsCreate():
				var job slavehandler.Job
				if err := json.Unmarshal(ev.Kv.Value, &job); err != nil {
					err = fmt.Errorf("watch job[%s] create json umarshal err: %s", string(ev.Kv.Key), err.Error())
					continue
				}
				srv.Jobs[job.ID] = &job
				job.InitNodeInfo(model.JobStatusAssigned, srv.UUID, srv.Hostname, srv.IP)
				srv.addJob(&job)
			case ev.IsModify():
				var job slavehandler.Job
				if err := json.Unmarshal(ev.Kv.Value, &job); err != nil {
					err = fmt.Errorf("watch job[%s] modify json umarshal err: %s", string(ev.Kv.Key), err.Error())
					continue
				}
				job.InitNodeInfo(model.JobStatusAssigned, srv.UUID, srv.Hostname, srv.IP)
				srv.modifyJob(&job)
			case ev.Type == mvccpb.DELETE:
				srv.deleteJob(slavehandler.GetJobIDFromKey(string(ev.Kv.Key)))
			default:
				logger.GetLogger().Warn(fmt.Sprintf("watch job unknown event type[%v] from job[%s]", ev.Type, string(ev.Kv.Key)))
			}
		}
	}
}

func (srv *NodeServer) watchSystemInfo() {
	rch := slavehandler.WatchSystem(srv.UUID)
	for wresp := range rch {
		for _, ev := range wresp.Events {
			switch {
			case ev.IsCreate() || ev.IsModify():
				key := string(ev.Kv.Key)
				if string(ev.Kv.Value) != model.NodeSystemInfoSwitch || srv.Node.UUID != getUUID(key) {
					logger.GetLogger().Error(fmt.Sprintf("get system info from node[%s] ,switch is not alive ", srv.UUID))
					continue
				}
				s, err := utils.GetServerInfo()
				if err != nil {
					logger.GetLogger().Error(fmt.Sprintf("get system info from node[%s] error: %s", srv.UUID, err.Error()))
					continue
				}
				b, err := json.Marshal(s)
				if err != nil {
					logger.GetLogger().Error(fmt.Sprintf("get system info from node[%s] json marshal error: %s", srv.UUID, err.Error()))
					continue
				}
				_, err = etcdconn.PutWithTtl(fmt.Sprintf(etcdconn.EtcdSystemGetKey, getUUID(key)), string(b), 5*60)
				if err != nil {
					logger.GetLogger().Error(fmt.Sprintf("get system info from node[%s] etcd put error: %s", srv.UUID, err.Error()))
					continue
				}

			}
		}
	}
}

func getUUID(key string) string {
	// /hades/node/<node_uuid>
	index := strings.LastIndex(key, "/")
	if index == -1 {
		return ""
	}
	return key[index+1:]
}

func (srv *NodeServer) watchOnce() {
	rch := slavehandler.WatchOnce()
	for wresp := range rch {
		for _, ev := range wresp.Events {
			switch {
			case ev.IsModify(), ev.IsCreate():
				// is not executed on this node
				if len(ev.Kv.Value) != 0 && string(ev.Kv.Value) != srv.UUID {
					continue
				}
				j, ok := srv.Jobs[slavehandler.GetJobIDFromKey(string(ev.Kv.Key))]
				if !ok {
					continue
				}
				go j.RunWithRecovery()
			}
		}
	}
}

func (srv *NodeServer) loadJobs() (err error) {
	defer func() {
		//When an outage occurs, get the context passed by panic and print it
		if r := recover(); r != nil {
			logger.GetLogger().Error(fmt.Sprintf("load jobs panic:%v", r))
		}
	}()
	// Obtain the scheduled job assigned by the local server
	jobs, err := slavehandler.GetJobs(srv.UUID)
	if err != nil {
		return
	}

	if len(jobs) == 0 {
		return
	}
	srv.Jobs = jobs

	for _, j := range jobs {
		j.InitNodeInfo(model.JobStatusAssigned, srv.UUID, srv.Hostname, srv.IP)
		srv.addJob(j)
	}

	return
}

func (srv *NodeServer) addJob(j *slavehandler.Job) {
	if err := j.Check(); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("job[%d] check error :%s", j.ID, err.Error()))
		return
	}
	taskFunc := slavehandler.CreateJob(j)
	if taskFunc == nil {
		logger.GetLogger().Error(fmt.Sprintf("Failed to create a task to process the Job. The task protocol was not supported%v", j.Type))
		return
	}
	err := goutil.PanicToError(func() {
		srv.Cron.AddFunc(j.Spec, taskFunc, srv.jobCronName(j.ID))
	})
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("Failed to add a task to the scheduler#%v", err.Error()))
	}

	return
}

func (srv *NodeServer) jobCronName(jobId int) string {
	return fmt.Sprintf(srv.UUID+"/%d", jobId)
}
