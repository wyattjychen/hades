package slaveservice

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"syscall"
	"time"

	"github.com/jakecoffman/cron"
	"github.com/wyattjychen/hades/internal/pkg/etcdconn"
	"github.com/wyattjychen/hades/internal/pkg/logger"
	"github.com/wyattjychen/hades/internal/pkg/model"
	"github.com/wyattjychen/hades/internal/pkg/mysqlconn"
	"github.com/wyattjychen/hades/internal/pkg/slave/slavehandler"
)

type NodeServer struct {
	*etcdconn.ServerReg
	*model.Node
	*cron.Cron

	Jobs slavehandler.Jobs
}

// Register into ETCD with /hades/node/<node_id>
func (srv *NodeServer) Register() error {
	pid, err := srv.exist(srv.UUID)
	if err != nil {
		return err
	}
	if pid != -1 {
		return fmt.Errorf("node[%s] with pid[%d] exist", srv.UUID, pid)
	}
	b, err := json.Marshal(&srv.Node)
	if err != nil {
		return fmt.Errorf("node[%s] with pid[%d] json error%s", srv.UUID, pid, err.Error())
	}
	//creates a new lease
	if err := srv.ServerReg.Register(fmt.Sprintf(etcdconn.EtcdNodeKey, srv.UUID), string(b)); err != nil {
		return err
	}
	return nil
}

// Check whether the node is registered with ETCD
// If yes, PID is returned. If no, -1 is returned
func (srv *NodeServer) exist(nodeUUID string) (pid int, err error) {
	resp, err := etcdconn.Get(fmt.Sprintf(etcdconn.EtcdNodeKey, nodeUUID))
	if err != nil {
		return
	}

	if len(resp.Kvs) == 0 {
		return -1, nil
	}

	if pid, err = strconv.Atoi(string(resp.Kvs[0].Value)); err != nil {
		if _, err = etcdconn.Delete(fmt.Sprintf(etcdconn.EtcdNodeKey, nodeUUID)); err != nil {
			return
		}
		return -1, nil
	}

	p, err := os.FindProcess(pid)
	if err != nil {
		return -1, nil
	}

	if p != nil && p.Signal(syscall.Signal(0)) == nil {
		return
	}
	return -1, nil
}

func (srv *NodeServer) Run() (err error) {
	defer func() {
		if err != nil {
			srv.Stop(err)
		}
	}()

	if err = srv.loadJobs(); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("node[%s] failed to load job error:%s", srv.UUID, err.Error()))
		return
	}
	insertId, err := srv.Node.Insert()
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("failed to create node[%s] into db error:%s", srv.UUID, err.Error()))
		return
	}
	srv.Node.ID = insertId
	//start cron
	srv.Cron.Start()
	go srv.watchJobs()
	//go srv.watchKilledProc()
	go srv.watchOnce()
	go srv.watchSystemInfo()
	return
}

func (srv *NodeServer) Stop(i interface{}) {
	srv.Down()
	_, err := etcdconn.Delete(fmt.Sprintf(etcdconn.EtcdNodeKey, srv.UUID))
	if err != nil {
		logger.GetLogger().Warn(fmt.Sprintf("failed to delete etcd node[%s] key error:%s", srv.UUID, err.Error()))
	}
	_, err = etcdconn.Delete(fmt.Sprintf(etcdconn.EtcdSystemGetKey, srv.UUID))
	if err != nil {
		logger.GetLogger().Warn(fmt.Sprintf("failed to delete system etcd node[%s] key error:%s", srv.UUID, err.Error()))
	}

	_ = srv.Client.Close()
	srv.Cron.Stop()
}

// Remove the survival information from mysql after the node instance is deactivated
func (srv *NodeServer) Down() {
	srv.Status = model.NodeConnFail
	srv.DownTime = time.Now().Unix()
	err := srv.Node.Update()
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("failed to update  node[%s] down  error:%s", srv.UUID, err.Error()))
	}
	err = mysqlconn.GetMysqlDB().Table(model.HadesJobTableName).Select("status").Where("run_on = ? ", srv.UUID).Updates(model.Job{
		Status: model.JobStatusNotAssigned,
	}).Error

	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("failed to update job on node[%s] down  error:%s", srv.UUID, err.Error()))
	}
}
