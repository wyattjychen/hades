package model

import (
	"fmt"

	"github.com/wyattjychen/hades/internal/pkg/mysqlconn"
)

const (
	NodeConnSuccess = 1
	NodeConnFail    = 2

	NodeSystemInfoSwitch = "alive"
)

// register to /hades/node/<node_uuid>/
type Node struct {
	ID       int    `json:"id" gorm:"column:id;primary_key;auto_increment"`
	PID      string `json:"pid" gorm:"size:16;column:pid;not null"`
	IP       string `json:"ip" gorm:"size:32;column:ip;default:''"`
	Hostname string `json:"hostname" gorm:"size:64;column:hostname;default:''"`
	UUID     string `json:"uuid" gorm:"size:128;column:uuid;not null;index:idx_node_uuid;"`
	Version  string `json:"version" gorm:"size:64;column:version;default:''"`
	Status   int    `json:"status" gorm:"size:1;column:status"`

	UpTime   int64 `json:"up" gorm:"column:up;not null"`
	DownTime int64 `json:"down" gorm:"column:down;default:0"`
}

func (n *Node) FindByUUID() error {
	return mysqlconn.GetMysqlDB().Table(HadesNodeTableName).Where("uuid = ? ", n.UUID).First(n).Error
}

func (n *Node) Delete() error {
	return mysqlconn.GetMysqlDB().Exec(fmt.Sprintf("delete from %s where uuid = ?", HadesNodeTableName), n.UUID).Error
}
