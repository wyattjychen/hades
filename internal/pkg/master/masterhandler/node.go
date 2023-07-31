package masterhandler

import (
	"fmt"

	"github.com/coreos/etcd/clientv3"
	"github.com/gin-gonic/gin"
	"github.com/wyattjychen/hades/internal/pkg/etcdconn"
	"github.com/wyattjychen/hades/internal/pkg/logger"
	"github.com/wyattjychen/hades/internal/pkg/master/mastermodel/masterrequest"
	"github.com/wyattjychen/hades/internal/pkg/master/mastermodel/masterresponse"
	"github.com/wyattjychen/hades/internal/pkg/master/masterservice"
	"github.com/wyattjychen/hades/internal/pkg/model"
)

type NodeRouter struct{}

var defaultNodeRouter = new(NodeRouter)

func (n *NodeRouter) Delete(c *gin.Context) {
	var req masterrequest.ByUUID
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[delete_node] masterrequest parameter error:%s", err.Error()))
		masterresponse.FailWithMessage(masterresponse.ErrorRequestParameter, "[delete_node] masterrequest parameter error", c)
		return
	}
	node := &model.Node{UUID: req.UUID}
	err := node.FindByUUID()
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[delete_node] find node by uuid :%s error:%s", req.UUID, err.Error()))
		masterresponse.FailWithMessage(masterresponse.ERROR, "[delete_node] db find error", c)
		return
	}
	if node.Status == model.NodeConnSuccess {
		masterresponse.FailWithMessage(masterresponse.ERROR, "[delete_node] can't delete a node that is already alive ", c)
		return
	}
	_, _ = etcdconn.Delete(fmt.Sprintf(etcdconn.EtcdJobKeyPrefix, req.UUID), clientv3.WithPrefix())
	err = node.Delete()
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[delete_node] into db error:%s", err.Error()))
		masterresponse.FailWithMessage(masterresponse.ERROR, "[delete_node] db delete error", c)
		return
	}
	masterresponse.OkWithMessage("delete success", c)
}

func (n *NodeRouter) Search(c *gin.Context) {
	var req masterrequest.ReqNodeSearch
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[search_node] masterrequest parameter error:%s", err.Error()))
		masterresponse.FailWithMessage(masterresponse.ErrorRequestParameter, "[search_node] masterrequest parameter error", c)
		return
	}
	req.Check()
	nodes, total, err := masterservice.DefaultNodeWatcher.Search(&req)
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[search_node] search node error:%s", err.Error()))
		masterresponse.FailWithMessage(masterresponse.ERROR, "[search_node] search node  error", c)
		return
	}
	var resultNodes []masterresponse.RspNodeSearch
	for _, node := range nodes {
		resultNode := masterresponse.RspNodeSearch{
			Node: node,
		}
		resultNode.JobCount, _ = masterservice.DefaultNodeWatcher.GetJobCount(node.UUID)
		resultNodes = append(resultNodes, resultNode)
	}

	masterresponse.OkWithDetailed(masterresponse.PageResult{
		List:     resultNodes,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, "search success", c)
}
