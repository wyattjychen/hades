package slavehandler

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/wyattjychen/hades/internal/pkg/logger"
	"github.com/wyattjychen/hades/internal/pkg/model"
)

type CMDHandler struct {
}

func (c *CMDHandler) Run(job *Job) (result string, err error) {
	var (
		cmd  *exec.Cmd
		proc *JobProc
	)
	if job.Timeout > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(job.Timeout)*time.Second)
		defer cancel()
		cmd = exec.CommandContext(ctx, job.Cmd[0], job.Cmd[1:]...)
	} else {
		cmd = exec.Command(job.Cmd[0], job.Cmd[1:]...)
	}

	// fmt.Println(job.Cmd[0])
	// fmt.Println(job.Cmd[1])

	var b bytes.Buffer
	cmd.Stdout = &b
	cmd.Stderr = &b

	err = cmd.Start()
	result = b.String()
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("%s\n%s", b.String(), err.Error()))
		return
	}
	proc = &JobProc{
		JobProc: &model.JobProc{
			ID:       cmd.Process.Pid,
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
	if err = cmd.Wait(); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("%s%s", b.String(), err.Error()))
		return b.String(), err
	}
	return b.String(), nil
}
