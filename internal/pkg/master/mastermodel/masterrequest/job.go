package masterrequest

import (
	"encoding/json"

	"github.com/wyattjychen/hades/internal/pkg/model"
)

type (
	ReqJobSearch struct {
		PageInfo
		ID     int           `json:"id" form:"id"`
		Name   string        `json:"name" form:"name"`
		RunOn  string        `json:"run_on" form:"run_on"`
		Type   model.JobType `json:"job_type" form:"type"`
		Status int           `json:"status" form:"status"`
	}
	ReqJobLogSearch struct {
		PageInfo
		Name     string `json:"name" form:"name"`
		JobId    int    `json:"job_id" form:"job_id"`
		NodeUUID string `json:"node_uuid" form:"node_uuid"`
		Success  *bool  `json:"success" form:"success"`
	}
	ReqJobUpdate struct {
		*model.Job
		Allocation int `json:"allocation" form:"allocation" binding:"required"`
	}
)

func (r *ReqJobUpdate) Valid() error {
	if r.Allocation == 0 {
		r.Allocation = model.AutoAllocation
	}
	notifyTo, _ := json.Marshal(r.NotifyToArray)
	r.NotifyTo = notifyTo
	return r.Check()
}
