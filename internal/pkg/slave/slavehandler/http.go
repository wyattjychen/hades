package slavehandler

import (
	"strings"
	"time"

	"github.com/wyattjychen/hades/internal/pkg/httpconn"
	"github.com/wyattjychen/hades/internal/pkg/model"
)

type HTTPHandler struct {
}

const HttpExecTimeout = 300

func (h *HTTPHandler) Run(job *Job) (result string, err error) {
	var (
		proc *JobProc
	)
	proc = &JobProc{
		JobProc: &model.JobProc{
			ID:       0,
			JobID:    job.ID,
			NodeUUID: job.RunOn,
			JobProcVal: model.JobProcVal{
				Time: time.Now(),
			},
		},
	}

	err = proc.Start()
	if err != nil {
		return
	}
	defer proc.Stop()
	if job.Timeout <= 0 || job.Timeout > HttpExecTimeout {
		job.Timeout = HttpExecTimeout
	}
	if job.HttpMethod == model.HTTPMethodGet {
		//logger.GetLogger().Debug(fmt.Sprintf("Job Command:%s   Timeout:%d",job.Command, job.Timeout))
		result, err = httpconn.Get(job.Command, job.Timeout)
	} else {
		urlFields := strings.Split(job.Command, "?")
		url := urlFields[0]
		var body string
		if len(urlFields) >= 2 {
			body = urlFields[1]
		}
		result, err = httpconn.PostJson(url, body, job.Timeout)
	}
	return
}
