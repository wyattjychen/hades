package hadeserrors

import "errors"

var (
	EtcdNotInitErr = errors.New("etcd is not initialized")

	EmptyJobNameErr    = errors.New("Name of job is empty.")
	EmptyJobCommandErr = errors.New("Command of job is empty.")
)
