package model

import (
	"encoding/json"

	"github.com/wyattjychen/hades/internal/pkg/mysqlconn"
)

type JobType int

const (
	JobTypeCmd  = JobType(1)
	JobTypeHttp = JobType(2)

	HTTPMethodGet  = 1
	HTTPMethodPost = 2

	JobExcSuccess = 1
	JobExcFail    = 0

	JobStatusNotAssigned = 0
	JobStatusAssigned    = 1

	ManualAllocation = 1
	AutoAllocation   = 2
)

// register to  /hades/job/<node_uuid>/<job_id>
type Job struct {
	ID      int    `json:"id" gorm:"column:id;primary_key;auto_increment"`
	Name    string `json:"name" gorm:"size:64;column:name;not null;index:idx_job_name" binding:"required"`
	Command string `json:"command" gorm:"type:text;column:command;not null" binding:"required"`
	//preset script ID
	ScriptID      []byte `json:"-"  gorm:"size:256;column:script_id;default:null"`
	ScriptIDArray []int  `json:"script_id" gorm:"-"`
	//Timeout setting of job execution time, which is effective when it is greater than 0.
	Timeout int64 `json:"timeout" gorm:"size:13;column:timeout;default:0"`
	// Retry times of task execution failures
	// The default value is 0
	RetryTimes int `json:"retry_times" gorm:"size:4,column:retry_times;default:0"`
	// Retry interval for task execution failure
	// in seconds. If the value is less than 0, try again immediately
	RetryInterval int64   `json:"retry_interval" gorm:"size:10;column:retry_interval;default:0"`
	Type          JobType `json:"job_type" gorm:"size:1;column:type;not null;" binding:"required"`
	HttpMethod    int     `json:"http_method" gorm:"size:1;column:http_method"`
	NotifyType    int     `json:"notify_type" gorm:"size:1;column:notify_type;not null"`
	// Whether to allocate nodes
	Status        int    `json:"status" gorm:"size:1;column:status;not null;default:0;index:idx_job_status"`
	NotifyTo      []byte `json:"-" gorm:"size:256;column:notify_to;default:null"`
	NotifyToArray []int  `json:"notify_to" gorm:"-"`
	Spec          string `json:"spec" gorm:"size:64;column:spec;not null"`
	RunOn         string `json:"run_on" gorm:"size:128;column:run_on;index:idx_job_run_on;"`
	Note          string `json:"note" gorm:"size:512;column:note;default:''"`
	Created       int64  `json:"created" gorm:"column:created;not null"`
	Updated       int64  `json:"updated" gorm:"column:updated;default:0"`

	Hostname string   `json:"host_name" gorm:"-"`
	Ip       string   `json:"ip" gorm:"-"`
	Cmd      []string `json:"cmd" gorm:"-"`
}

func (j *Job) Unmarshal() (err error) {
	err = json.Unmarshal(j.NotifyTo, &j.NotifyToArray)
	err = json.Unmarshal(j.ScriptID, &j.ScriptIDArray)
	return
}

func (j *Job) InitNodeInfo(status int, nodeUUID, hostname, ip string) {
	j.Status, j.RunOn, j.Hostname, j.Ip = status, nodeUUID, hostname, ip
}

func (j *Job) Update() error {
	return mysqlconn.GetMysqlDB().Table(HadesJobTableName).Updates(j).Error
}
