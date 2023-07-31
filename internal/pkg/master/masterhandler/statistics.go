package masterhandler

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wyattjychen/hades/internal/pkg/etcdconn"
	"github.com/wyattjychen/hades/internal/pkg/logger"
	"github.com/wyattjychen/hades/internal/pkg/master/mastermodel/masterrequest"
	"github.com/wyattjychen/hades/internal/pkg/master/mastermodel/masterresponse"
	"github.com/wyattjychen/hades/internal/pkg/master/masterservice"
	"github.com/wyattjychen/hades/internal/pkg/model"
	"github.com/wyattjychen/hades/internal/pkg/utils"
)

type StatRouter struct {
}

var defaultStatRouter = new(StatRouter)

func (s *StatRouter) GetTodayStatistics(c *gin.Context) {
	jobExcSuccess, err := masterservice.DefaultJobService.GetTodayJobExcCount(model.JobExcSuccess)
	if err != nil {
		logger.GetLogger().Warn(fmt.Sprintf("[get_statisitcs] GetTodayJobExcCount(successs)  error:%s", err.Error()))
	}
	jobExcFail, err := masterservice.DefaultJobService.GetTodayJobExcCount(model.JobExcFail)
	if err != nil {
		logger.GetLogger().Warn(fmt.Sprintf("[get_statisitcs] GetTodayJobExcCount(fail) error:%s", err.Error()))
	}
	jobRunningCount, err := masterservice.DefaultJobService.GetRunningJobCount()
	if err != nil {
		logger.GetLogger().Warn(fmt.Sprintf("[get_statisitcs] GetRunningJobCount error:%s", err.Error()))
	}
	normalNodeCount, err := masterservice.DefaultNodeWatcher.GetNodeCount(model.NodeConnSuccess)
	if err != nil {
		logger.GetLogger().Warn(fmt.Sprintf("[get_statisitcs] GetNodeCount(success) error:%s", err.Error()))
	}
	failNodeCount, err := masterservice.DefaultNodeWatcher.GetNodeCount(model.NodeConnFail)
	if err != nil {
		logger.GetLogger().Warn(fmt.Sprintf("[get_statisitcs] GetNodeCount(fail) error:%s", err.Error()))
	}
	masterresponse.OkWithDetailed(masterresponse.RspSystemStatistics{
		NormalNodeCount:    normalNodeCount,
		FailNodeCount:      failNodeCount,
		JobExcSuccessCount: jobExcSuccess,
		JobRunningCount:    jobRunningCount,
		JobExcFailCount:    jobExcFail,
	}, "ok", c)
}

func (s *StatRouter) GetWeekStatistics(c *gin.Context) {
	t := time.Now()
	jobExcSuccess, err := masterservice.DefaultJobService.GetJobExcCount(t.Unix()-60*60*24*7, t.Unix(), model.JobExcSuccess)
	if err != nil {
		logger.GetLogger().Warn(fmt.Sprintf("[get_week_statisitcs] GetTodayJobExcCount(successs)  error:%s", err.Error()))
	}
	jobExcFail, err := masterservice.DefaultJobService.GetJobExcCount(t.Unix()-60*60*24*7, t.Unix(), model.JobExcFail)
	if err != nil {
		logger.GetLogger().Warn(fmt.Sprintf("[get_week_statisitcs] GetTodayJobExcCount(fail) error:%s", err.Error()))
	}
	masterresponse.OkWithDetailed(masterresponse.RspDateCount{
		SuccessDateCount: jobExcSuccess,
		FailDateCount:    jobExcFail,
	}, "ok", c)
}

func (s *StatRouter) GetSystemInfo(c *gin.Context) {
	var req masterrequest.ByUUID
	if err := c.ShouldBindQuery(&req); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[get_system_info] masterrequest parameter error:%s", err.Error()))
		masterresponse.FailWithMessage(masterresponse.ErrorRequestParameter, "[get_system_info] masterrequest parameter error", c)
		return
	}
	var server *utils.Server
	var err error
	if req.UUID == "" {
		//Get native information of admin
		server, err = utils.GetServerInfo()
		if err != nil {
			logger.GetLogger().Warn(fmt.Sprintf("[get_system_info]  error:%s", err.Error()))
			masterresponse.FailWithMessage(masterresponse.ERROR, "[get_system_info]  error", c)
			return
		}
	} else {
		//Set the survival time to 30 seconds
		_, err := etcdconn.PutWithTtl(fmt.Sprintf(etcdconn.EtcdSystemSwitchKey, req.UUID), model.NodeSystemInfoSwitch, 30)
		if err != nil {
			logger.GetLogger().Error(fmt.Sprintf("get system info from node[%s] etcd put error: %s", req.UUID, err.Error()))
			masterresponse.FailWithMessage(masterresponse.ERROR, "[get_system_info]  error", c)
			return
		}
		//There will be a delay. By default, wait 2s.
		time.Sleep(2 * time.Second)
		server, err = masterservice.GetNodeSystemInfo(req.UUID)
		if err != nil {
			logger.GetLogger().Error(fmt.Sprintf("get system info from node[%s] watch key error: %s", req.UUID, err.Error()))
			masterresponse.FailWithMessage(masterresponse.ERROR, "[get_system_info]  error", c)
			return
		}
	}
	masterresponse.OkWithDetailed(server, "ok", c)

}
