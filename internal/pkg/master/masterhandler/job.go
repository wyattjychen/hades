package masterhandler

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wyattjychen/hades/internal/pkg/config"
	"github.com/wyattjychen/hades/internal/pkg/etcdconn"
	"github.com/wyattjychen/hades/internal/pkg/logger"
	"github.com/wyattjychen/hades/internal/pkg/master/mastermodel/masterrequest"
	"github.com/wyattjychen/hades/internal/pkg/master/mastermodel/masterresponse"
	"github.com/wyattjychen/hades/internal/pkg/master/masterservice"
	"github.com/wyattjychen/hades/internal/pkg/model"
)

type JobRouter struct {
}

var defaultJobRouter = new(JobRouter)

func (j *JobRouter) CreateOrUpdate(c *gin.Context) {
	var req masterrequest.ReqJobUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[create_job] request parameter error:%s", err.Error()))
		masterresponse.FailWithMessage(masterresponse.ErrorRequestParameter, "[create_job] request parameter error", c)
		return
	}
	if err := req.Valid(); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("create_job check error:%s", err.Error()))
		masterresponse.FailWithMessage(masterresponse.ErrorJobFormat, "[create_job] check error", c)
		return
	}

	var err error
	var insertId int
	t := time.Now()

	if req.Allocation == model.AutoAllocation {
		if !config.GetConfig().System.CmdAutoAllocation && req.Type == model.JobTypeCmd {
			masterresponse.FailWithMessage(masterresponse.ERROR, "[create_job] The shell command is not supported to automatically assign nodes by default.", c)
			return
		}
		// Automatic allocation
		nodeUUID := masterservice.DefaultJobService.AutoAllocateNode()
		if nodeUUID == "" {
			logger.GetLogger().Error(fmt.Sprintf("[create_job] auto allocate node error"))
			masterresponse.FailWithMessage(masterresponse.ERROR, "[create_job] auto allocate node error", c)
			return
		}
		req.RunOn = nodeUUID
	} else if req.Allocation == model.ManualAllocation {
		// Manual assignment
		if len(req.RunOn) == 0 {
			masterresponse.FailWithMessage(masterresponse.ERROR, "[create_job] manually assigned node can't be null", c)
			return
		}
		node := &model.Node{UUID: req.RunOn}
		_ = node.FindByUUID()
		if node.Status == model.NodeConnFail {
			masterresponse.FailWithMessage(masterresponse.ERROR, "[create_job] manually assigned node inactivation", c)
			return
		}
	}
	if req.ID > 0 {
		//update
		job := &model.Job{ID: req.ID}
		_ = job.FindById()
		oldNodeUUID := job.RunOn
		if oldNodeUUID != "" {
			_, err = etcdconn.Delete(fmt.Sprintf(etcdconn.EtcdJobKey, oldNodeUUID, req.ID))
			if err != nil {
				logger.GetLogger().Error(fmt.Sprintf("[update_job] delete etcd node[%s]  error:%s", oldNodeUUID, err.Error()))
				masterresponse.FailWithMessage(masterresponse.ERROR, "[update_job] delete etcd node error", c)
				return
			}
		}
		req.Updated = t.Unix()
		err = req.Update()
		if err != nil {
			logger.GetLogger().Error(fmt.Sprintf("[update_job] into db  error:%s", err.Error()))
			masterresponse.FailWithMessage(masterresponse.ERROR, "[update_job] into db id error", c)
			return
		}
	} else {
		//create
		req.Created = t.Unix()
		insertId, err = req.Insert()
		if err != nil {
			logger.GetLogger().Error(fmt.Sprintf("[create_job] insert job into db error:%s", err.Error()))
			masterresponse.FailWithMessage(masterresponse.ERROR, "[create_job] insert job into db error", c)
			return
		}
		req.ID = insertId
	}
	b, err := json.Marshal(req)
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[create_job] json marshal job error:%s", err.Error()))
		masterresponse.FailWithMessage(masterresponse.ERROR, "[create_job] json marshal job error", c)
		return
	}
	_, err = etcdconn.Put(fmt.Sprintf(etcdconn.EtcdJobKey, req.RunOn, req.ID), string(b))
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[create_job] etcd put job error:%s", err.Error()))
		masterresponse.FailWithMessage(masterresponse.ERROR, "[create_job] etcd put job error", c)
		return
	}

	masterresponse.OkWithDetailed(req, "operate success", c)
}
