package etcdconn

const (
	EtcdKeyPrefix = "/hades/"

	//hades/node/<node_uuid>
	EtcdNodeKeyPrefix = EtcdKeyPrefix + "node/"
	EtcdNodeKey       = EtcdNodeKeyPrefix + "%s"

	//hades/proc/<node_uuid>/<job_id>/<pid>
	EtcdProcKeyPrefix     = EtcdKeyPrefix + "proc/"
	EtcdNodeProcKeyPrefix = EtcdProcKeyPrefix + "%s/"
	EtcdJobProcKeyPrefix  = EtcdNodeProcKeyPrefix + "%d/"
	EtcdProcKey           = EtcdJobProcKeyPrefix + "%d"

	//hades/once/<jobID>
	EtcdOnceKeyPrefix = EtcdKeyPrefix + "once/"
	EtcdOnceKey       = EtcdOnceKeyPrefix + "%d"

	//hades/job/<node_uuid>/<job_id>
	EtcdJobKeyPrefix = EtcdKeyPrefix + "job/%s/"
	EtcdJobKey       = EtcdJobKeyPrefix + "%d"

	//hades/system/<node_uuid>
	EtcdSystemKeyPrefix = EtcdKeyPrefix + "system/"
	EtcdSystemSwitchKey = EtcdSystemKeyPrefix + "switch/" + "%s"
	EtcdSystemGetKey    = EtcdSystemKeyPrefix + "get/" + "%s"
)
