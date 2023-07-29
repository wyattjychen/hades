package etcdconn

import (
	"context"
	"fmt"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/wyattjychen/hades/internal/pkg/config"
	"github.com/wyattjychen/hades/internal/pkg/logger"
	"github.com/wyattjychen/hades/internal/pkg/utils/hadeserrors"
)

var defalutEtcd *Client

type Client struct {
	*clientv3.Client
	reqTimeout time.Duration
}

func Init(endpoints []string, dialTimeout, reqTimeout int64) (*Client, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: time.Duration(dialTimeout) * time.Second,
	})
	if err != nil {
		fmt.Printf("connect to etcd failed, err:%v\n", err)
		return nil, err
	}
	defalutEtcd = &Client{
		Client:     cli,
		reqTimeout: time.Duration(reqTimeout) * time.Second,
	}
	return defalutEtcd, nil
}

func GetEtcdClient() *Client {
	if defalutEtcd == nil {
		logger.GetLogger().Error("etcd is not initialized")
		return nil
	}
	return defalutEtcd
}

func Get(key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
	if defalutEtcd == nil {
		return nil, hadeserrors.EtcdNotInitErr
	}
	ctx, cancel := NewEtcdTimeoutContext()
	defer cancel()
	return defalutEtcd.Get(ctx, key, opts...)
}

func Put(key, val string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error) {
	if defalutEtcd == nil {
		return nil, hadeserrors.EtcdNotInitErr
	}
	ctx, cancel := NewEtcdTimeoutContext()
	defer cancel()
	return defalutEtcd.Put(ctx, key, val, opts...)
}

func Delete(key string, opts ...clientv3.OpOption) (*clientv3.DeleteResponse, error) {
	if defalutEtcd == nil {
		return nil, hadeserrors.EtcdNotInitErr
	}
	ctx, cancel := NewEtcdTimeoutContext()
	defer cancel()
	return defalutEtcd.Delete(ctx, key, opts...)
}

type etcdTimeoutContext struct {
	context.Context
	etcdEndpoints []string
}

func NewEtcdTimeoutContext() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), defalutEtcd.reqTimeout)
	etcdCtx := &etcdTimeoutContext{}
	etcdCtx.Context = ctx
	etcdCtx.etcdEndpoints = config.GetConfig().Etcd.Endpoints
	return etcdCtx, cancel
}
