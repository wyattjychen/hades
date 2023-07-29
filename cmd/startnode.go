package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/wyattjychen/hades/internal/pkg/config"
	"github.com/wyattjychen/hades/internal/pkg/logger"
	"github.com/wyattjychen/hades/internal/pkg/master/masterhandler"
	"github.com/wyattjychen/hades/internal/pkg/master/masterservice"
	"github.com/wyattjychen/hades/internal/pkg/mysqlconn"
	"github.com/wyattjychen/hades/internal/pkg/server"
)

func Start() {
	nodeType, cfgFile, err := server.NewNode()
	if err != nil {
		fmt.Println("node start failed!")
		os.Exit(1)
	}
	if nodeType == "master" {
		srv, err := server.NewMasterServer(nodeType, cfgFile)
		if err != nil {
			logger.GetLogger().Error(fmt.Sprintf("create new master server error:%s", err.Error()))
			fmt.Println("master server start failed!")
			os.Exit(1)
		}
		fmt.Println(srv.Addr)
		// Register the API routing service
		srv.RegisterRouters(masterhandler.RegisterRouters)
		masterservice.DefaultNodeWatcher = masterservice.NewNodeWatcherService()
		err = masterservice.DefaultNodeWatcher.Watch()
		if err != nil {
			logger.GetLogger().Error(fmt.Sprintf("resolver  error:%#v", err))
		}
		//init db table
		err = masterservice.RegisterTables(mysqlconn.GetMysqlDB())
		if err != nil {
			logger.GetLogger().Error(fmt.Sprintf("init db table error:%#v", err))
		}
		// TODO: Notify operation

		// log cleaner
		var closeChan chan struct{}
		period := config.GetConfig().System.LogCleanPeriod
		if period > 0 {
			closeChan = masterservice.RunLogCleaner(time.Duration(period)*time.Minute, config.GetConfig().System.LogCleanExpiration)
		}
		err = srv.ListenAndServe()
		if err != nil {
			logger.GetLogger().Error(fmt.Sprintf("startup api server error:%v", err.Error()))
			close(closeChan)
			os.Exit(1)
		}
		os.Exit(0)

	} else {
		// Start slave
		fmt.Println("node start")
	}

}
