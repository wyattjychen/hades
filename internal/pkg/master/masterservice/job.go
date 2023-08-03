package masterservice

import (
	"fmt"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/wyattjychen/hades/internal/pkg/etcdconn"
	"github.com/wyattjychen/hades/internal/pkg/logger"
	"github.com/wyattjychen/hades/internal/pkg/master/mastermodel/masterrequest"
	"github.com/wyattjychen/hades/internal/pkg/master/mastermodel/masterresponse"
	"github.com/wyattjychen/hades/internal/pkg/master/masterservice/balance"
	"github.com/wyattjychen/hades/internal/pkg/model"
	"github.com/wyattjychen/hades/internal/pkg/mysqlconn"
	"github.com/wyattjychen/hades/internal/pkg/utils"
)

type JobService struct {
}

var DefaultJobService = new(JobService)

const MaxJobCount = 10000

func (j *JobService) GetNotAssignedJob() (jobs []model.Job, err error) {
	err = mysqlconn.GetMysqlDB().Table(model.HadesJobTableName).Where("status = ?", model.JobStatusNotAssigned).Find(&jobs).Error
	return
}

func (j *JobService) AutoAllocateNode(balanceType balance.BalanceType) string {
	//Get all the living nodes
	nodeList := DefaultNodeWatcher.List2Array()
	switch balanceType {
	case balance.RADOMBALANCE:
		resultNodeUUID, err := balance.DefaultRandomBalancer.DoBalance(nodeList)
		// logger.GetLogger().Info("random balance")
		if err != nil {
			return ""
		}
		return resultNodeUUID
	case balance.ROUNDROBINBALANCE:
		resultNodeUUID, err := balance.DefaultRoundRobinBalancer.DoBalance(nodeList)
		if err != nil {
			return ""
		}
		return resultNodeUUID
	case balance.SHUFFLEBALANCE:
		resultNodeUUID, err := balance.DefaultSuffleBalancer.DoBalance(nodeList)
		if err != nil {
			return ""
		}
		return resultNodeUUID
	case balance.CONSISTANTHASHBALANCE:
		resultNodeUUID, err := balance.DefaultConsistantHashBalancer.DoBalance(nodeList)
		if err != nil {
			return ""
		}
		return resultNodeUUID
	default:
		//default: Least exec of tasks preferred.
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

// Get the total number of tasks executed today 1 indicates success 0 indicates failure
func (j *JobService) GetTodayJobExcCount(success int) (int64, error) {
	db := mysqlconn.GetMysqlDB().Table(model.HadesJobLogTableName).Where("start_time > ? and end_time!=0 and success = ?", utils.GetTodayUnix(), success)
	var total int64
	err := db.Count(&total).Error
	if err != nil {
		return 0, err
	}
	return total, nil
}

func (j *JobService) GetRunningJobCount() (int64, error) {
	wresp, err := etcdconn.Get(fmt.Sprintf(etcdconn.EtcdProcKeyPrefix), clientv3.WithPrefix(), clientv3.WithCountOnly())
	if err != nil {
		return 0, err
	}

	return wresp.Count, nil
}

// The number of tasks per day in a certain period of time
func (j *JobService) GetJobExcCount(start, end int64, success int) ([]masterresponse.DateCount, error) {
	var dateCount []masterresponse.DateCount
	db := mysqlconn.GetMysqlDB().Table(model.HadesJobLogTableName).Select("FROM_UNIXTIME( start_time, '%Y-%m-%d' ) AS date", "COUNT( * ) AS count ").Group("date").Order("date ASC").Where("start_time > ? and start_time<?  and end_time!=0 and success = ?", start, end, success)
	err := db.Find(&dateCount).Error
	if err != nil {
		return nil, err
	}
	return dateCount, nil
}

func (j *JobService) Search(s *masterrequest.ReqJobSearch) ([]model.Job, int64, error) {
	db := mysqlconn.GetMysqlDB().Table(model.HadesJobTableName)
	if s.ID > 0 {
		db = db.Where("id = ?", s.ID)
	}
	if len(s.Name) > 0 {
		db = db.Where("name like ?", s.Name+"%")
	}
	if len(s.RunOn) > 0 {
		db.Where("run_on = ?", s.RunOn)
	}
	if s.Type > 0 {
		db.Where("type = ?", s.Type)
	}
	if s.Status > 0 {
		db.Where("status = ? ", s.Status)
	}
	jobs := make([]model.Job, 2)
	var total int64
	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	err = db.Limit(s.PageSize).Offset((s.Page - 1) * s.PageSize).Find(&jobs).Error
	if err != nil {
		return nil, 0, err
	}
	return jobs, total, nil
}

func (j *JobService) SearchJobLog(s *masterrequest.ReqJobLogSearch) ([]model.JobLog, int64, error) {
	db := mysqlconn.GetMysqlDB().Table(model.HadesJobLogTableName)
	if len(s.Name) > 0 {
		db = db.Where("name like ?", s.Name+"%")
	}
	if s.JobId > 0 {
		db.Where("job_id = ?", s.JobId)
	}
	if len(s.NodeUUID) > 0 {
		db.Where("node_uuid = ?", s.NodeUUID)
	}
	if s.Success != nil {
		logger.GetLogger().Info(fmt.Sprintf("%v", *s.Success))
		db.Where("success = ? ", *s.Success)
	}
	jobLogs := make([]model.JobLog, 2)
	var total int64
	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	err = db.Limit(s.PageSize).Offset((s.Page - 1) * s.PageSize).Order("start_time desc").Find(&jobLogs).Error
	if err != nil {
		return nil, 0, err
	}

	return jobLogs, total, nil
}
