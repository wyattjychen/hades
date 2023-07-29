package etcdconn

const (
	EtcdKeyPrefix = "/hades/"

	//key /hades/node/<node_uuid>
	EtcdNodeKeyPrefix = EtcdKeyPrefix + "node/"
	EtedNodeKey       = EtcdNodeKeyPrefix + "%s"

	//key /hades/job/<node_uuid>/<job_id>
	EtcdJobKeyPrefix = EtcdKeyPrefix + "job/%s/"
	EtcdJobKey       = EtcdJobKeyPrefix + "%d"
)
