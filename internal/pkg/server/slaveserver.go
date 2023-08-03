package server

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/jakecoffman/cron"
	"github.com/wyattjychen/hades/internal/pkg/config"
	"github.com/wyattjychen/hades/internal/pkg/etcdconn"
	"github.com/wyattjychen/hades/internal/pkg/logger"
	"github.com/wyattjychen/hades/internal/pkg/model"
	"github.com/wyattjychen/hades/internal/pkg/mysqlconn"
	"github.com/wyattjychen/hades/internal/pkg/notify"
	"github.com/wyattjychen/hades/internal/pkg/slave/slavehandler"
	"github.com/wyattjychen/hades/internal/pkg/slave/slaveservice"
	"github.com/wyattjychen/hades/internal/pkg/utils"
)

func NewSlaveServer(nodeType, configFile string) (*slaveservice.NodeServer, error) {

	if configFile == "" {
		configFile = "config.json"
	}
	defaultCfg, err := config.LoadConfig(nodeType, configFile)
	if err != nil {
		fmt.Printf("get config failed: %s", err.Error())
		return nil, err
	}
	logCfg := defaultCfg.Log
	mysqlCfg := defaultCfg.Mysql
	etcdCfg := defaultCfg.Etcd

	uuid, err := utils.UUID()
	if err != nil {
		return nil, err
	}

	logger.Init(nodeType+uuid, logCfg.Level, logCfg.Format, logCfg.Prefix, logCfg.Director, logCfg.ShowLine, logCfg.EncodeLevel, logCfg.StacktraceKey, logCfg.LogInConsole)

	// todo: notify init
	notify.Init(&notify.Mail{
		Port:     defaultCfg.Email.Port,
		From:     defaultCfg.Email.From,
		Host:     defaultCfg.Email.Host,
		Secret:   defaultCfg.Email.Secret,
		Nickname: defaultCfg.Email.Nickname,
	})

	// db-mysql
	dsn := mysqlCfg.NewDsn()
	createSql := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` DEFAULT CHARACTER SET utf8mb4 ;", mysqlCfg.Dbname)
	if err := mysqlconn.CreateDatabase(dsn, "mysql", createSql); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("create mysql database failed , error:%s", err.Error()))
		fmt.Println("failed")
		return nil, err
	}
	_, err = mysqlconn.Init(mysqlCfg.NowDsn(), mysqlCfg.LogMode, mysqlCfg.MaxIdleConns, mysqlCfg.MaxOpenConns)
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("slave server:init mysql failed , error:%s", err.Error()))
		fmt.Println(fmt.Sprintf("slave server:init mysql failed , error:%s", err.Error()))
		return nil, err
	} else {
		logger.GetLogger().Info("slave server:init mysql success")
		fmt.Println("slave server:init mysql success")
	}

	// db-etcd
	_, err = etcdconn.Init(etcdCfg.Endpoints, etcdCfg.DialTimeout, etcdCfg.ReqTimeout)
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("server:init etcd failed , error:%s", err.Error()))
		fmt.Println("server etcd-get failed")
		return nil, err
	} else {
		logger.GetLogger().Info("slave server:init etcd success.")
		fmt.Println("slave server init etcd success.")
	}

	ip, err := utils.LocalIP()
	if err != nil {
		return nil, err
	}
	hostname, err := os.Hostname()
	if err != nil {
		hostname = uuid
		err = nil
	}
	return &slaveservice.NodeServer{
		Node: &model.Node{
			UUID:     uuid,
			PID:      strconv.Itoa(os.Getpid()),
			IP:       ip.String(),
			Hostname: hostname,
			UpTime:   time.Now().Unix(),
			Status:   model.NodeConnSuccess,
			// Version:  config.GetConfig().System.Version,
		},
		Cron:      cron.New(),
		Jobs:      make(slavehandler.Jobs, 8),
		ServerReg: etcdconn.NewServerReg(config.GetConfig().System.NodeTtl),
	}, nil

}
