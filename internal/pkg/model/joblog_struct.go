package model

import (
	"fmt"

	"github.com/wyattjychen/hades/internal/pkg/mysqlconn"
)

type JobLog struct {
	ID       int    `json:"id" gorm:"column:id;primary_key;auto_increment"`
	Name     string `json:"name" gorm:"size:64;column:name;index:idx_job_log_name;not null"`
	JobId    int    `json:"job_id" gorm:"column:job_id;index:idx_job_log_id; not null"`
	Command  string `json:"command" gorm:"size:512;column:command"`
	IP       string `json:"ip" gorm:"size:32;column:ip"` // node ip
	Hostname string `json:"hostname" gorm:"size:32;column:hostname"`
	NodeUUID string `json:"node_uuid" gorm:"size:128;column:node_uuid;not null;index:idx_job_log_node"`
	Success  bool   `json:"success" gorm:"size:1;column:success;not null"`

	Output string `json:"output" gorm:"size:512;column:output;"`
	Spec   string `json:"spec" gorm:"size:64;column:spec;not null" `

	RetryTimes int   `json:"retry_times" gorm:"size:4;column:retry_times;default:0"`
	StartTime  int64 `json:"start_time" gorm:"column:start_time;not null;"`
	EndTime    int64 `json:"end_time" gorm:"column:end_time;default:0;"`
}

func (jb *JobLog) Update() error {
	return mysqlconn.GetMysqlDB().Table(HadesJobLogTableName).Updates(jb).Error
}

func (jb *JobLog) Delete() error {
	return mysqlconn.GetMysqlDB().Exec(fmt.Sprintf("delete from %s where id = ?", HadesJobLogTableName), jb.ID).Error
}

func (jb *JobLog) Insert() (insertId int, err error) {
	err = mysqlconn.GetMysqlDB().Table(HadesJobLogTableName).Create(jb).Error
	if err == nil {
		insertId = jb.ID
	}
	return
}

func (jb *JobLog) TableName() string {
	return HadesJobLogTableName
}
