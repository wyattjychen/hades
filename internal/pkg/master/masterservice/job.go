package masterservice

import (
	"fmt"
	"time"

	"github.com/wyattjychen/hades/internal/pkg/logger"
	"github.com/wyattjychen/hades/internal/pkg/model"
	"github.com/wyattjychen/hades/internal/pkg/mysqlconn"
)

type JobService struct {
}

var DefaultJobService = new(JobService)

const MaxJobCount = 10000

func (j *JobService) GetNotAssignedJob() (jobs []model.Job, err error) {
	err = mysqlconn.GetMysqlDB().Table(model.HadesJobTableName).Where("status = ?", model.JobStatusNotAssigned).Find(&jobs).Error
	return
}

// Give priority to the node with the least number of tasks
// TODO: add other load balance algorithms
func (j *JobService) AutoAllocateNode() string {
	//Get all the living nodes
	nodeList := DefaultNodeWatcher.List2Array()
	resultCount, resultNodeUUID := MaxJobCount, ""
	for _, nodeUUID := range nodeList {
		//Check the database to see if it is alive
		node := &model.Node{UUID: nodeUUID}
		err := node.FindByUUID()
		if err != nil {
			continue
		}
		if node.Status == model.NodeConnFail {
			//The node has failed
			delete(DefaultNodeWatcher.nodeList, nodeUUID)
			continue
		}
		count, err := DefaultNodeWatcher.GetJobCount(nodeUUID)
		if err != nil {
			logger.GetLogger().Warn(fmt.Sprintf("node[%s] get job conut error:%s", nodeUUID, err.Error()))
			continue
		}
		if resultCount > count {
			resultCount, resultNodeUUID = count, nodeUUID
		}
	}
	return resultNodeUUID
}

func RunLogCleaner(cleanPeriod time.Duration, expiration int64) (close chan struct{}) {
	t := time.NewTicker(cleanPeriod)
	close = make(chan struct{})
	go func() {
		for {
			select {
			case <-t.C:
				err := cleanupLogs(expiration)
				if err != nil {
					logger.GetLogger().Error(fmt.Sprintf("clean up logs at time:%v error:%s", time.Now(), err.Error()))
				}
			case <-close:
				t.Stop()
				return
			}
		}
	}()
	return
}

func cleanupLogs(expirationTime int64) error {
	sql := fmt.Sprintf("delete from %s where start_time < ?", model.HadesJobLogTableName)
	return mysqlconn.GetMysqlDB().Exec(sql, time.Now().Unix()-expirationTime).Error
}
