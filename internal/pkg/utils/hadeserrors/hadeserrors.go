package hadeserrors

import "errors"

var (
	ClientNotFoundErr   = errors.New("mysql client not found")
	ClientDbNameNullErr = errors.New("mysql dbname is null")
	ValueMayChangedErr  = errors.New("The value has been changed by others on this time.")
	EtcdNotInitErr      = errors.New("etcd is not initialized")

	NotFoundErr = errors.New("Record not found.")

	EmptyJobNameErr        = errors.New("Name of job is empty.")
	EmptyJobCommandErr     = errors.New("Command of job is empty.")
	IllegalJobIdErr        = errors.New("Invalid id that includes illegal characters such as '/' '\\'.")
	IllegalJobGroupNameErr = errors.New("Invalid job group name that includes illegal characters such as '/' '\\'.")

	EmptyScriptNameErr    = errors.New("Name of script is empty.")
	EmptyScriptCommandErr = errors.New("Command of script is empty.")
	EmptyNodeGroupNameErr = errors.New("Name of node group is empty.")
	IllegalNodeGroupIdErr = errors.New("Invalid node group id that includes illegal characters such as '/'.")
)
