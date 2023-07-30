package slavehandler

import (
	"fmt"

	"github.com/coreos/etcd/clientv3"
	"github.com/wyattjychen/hades/internal/pkg/etcdconn"
	"github.com/wyattjychen/hades/internal/pkg/model"
)

type Handler interface {
	Run(job *Job) (string, error)
}

func CreateHandler(j *Job) Handler {
	var handler Handler = nil
	switch j.Type {
	case model.JobTypeCmd:
		handler = new(CMDHandler)
	case model.JobTypeHttp:
		handler = new(HTTPHandler)
	}
	return handler
}

func WatchSystem(nodeUUID string) clientv3.WatchChan {
	return etcdconn.Watch(fmt.Sprintf(etcdconn.EtcdSystemSwitchKey, nodeUUID), clientv3.WithPrefix())
}

// Execute the job immediately
// value
// If a single node is executed, the value is nodeUUID
// If the node where the job is located needs to be executed, the value is null ""
func WatchOnce() clientv3.WatchChan {
	return etcdconn.Watch(etcdconn.EtcdOnceKeyPrefix, clientv3.WithPrefix())
}
