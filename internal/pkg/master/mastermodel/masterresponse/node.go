package masterresponse

import "github.com/wyattjychen/hades/internal/pkg/model"

type (
	RspNodeSearch struct {
		model.Node
		JobCount int `json:"job_count"`
	}
)
