package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wyattjychen/hades/internal/pkg/config"
	"github.com/wyattjychen/hades/internal/pkg/etcdconn"
	"github.com/wyattjychen/hades/internal/pkg/logger"
	"github.com/wyattjychen/hades/internal/pkg/mysqlconn"
	"github.com/wyattjychen/hades/internal/pkg/notify"
)

var NodeConfigOps = NodeConfigOptions{}

type Server struct {
	Engine     *gin.Engine
	HttpServer *http.Server
	Addr       string
	Routers    []func(*gin.Engine)
}

func (srv *Server) setupSignal() {
	go func() {
		var sigChan = make(chan os.Signal, 1)
		signal.Notify(sigChan /*syscall.SIGUSR1,*/, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM)
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownMaxAge)
		defer shutdownCancel()

		for sig := range sigChan {
			if sig == syscall.SIGINT || sig == syscall.SIGHUP || sig == syscall.SIGTERM {
				logger.GetLogger().Error(fmt.Sprintf("Graceful shutdown:signal %v to stop api-server ", sig))
				fmt.Println("shut down now.")
				srv.Shutdown(shutdownCtx)
			} else {
				logger.GetLogger().Info(fmt.Sprintf("Caught signal %v", sig))
			}
		}
		logger.Shutdown()
	}()
}

func (srv *Server) Shutdown(ctx context.Context) {
	select {
	case <-time.After(shutdownWait):
	}
	srv.HttpServer.Shutdown(ctx)
}

func NewMasterServer(nodeType, configFile string) (*Server, error) {
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

	logger.Init(nodeType, logCfg.Level, logCfg.Format, logCfg.Prefix, logCfg.Director, logCfg.ShowLine, logCfg.EncodeLevel, logCfg.StacktraceKey, logCfg.LogInConsole)

	//notify.
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
		logger.GetLogger().Error(fmt.Sprintf("server:init mysql failed , error:%s", err.Error()))
		fmt.Println(fmt.Sprintf("server:init mysql failed , error:%s", err.Error()))
		return nil, err
	} else {
		logger.GetLogger().Info("server:init mysql success")
		fmt.Println("server:init mysql success")
	}

	// db-etcd
	_, err = etcdconn.Init(etcdCfg.Endpoints, etcdCfg.DialTimeout, etcdCfg.ReqTimeout)
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("server:init etcd failed , error:%s", err.Error()))
		fmt.Println("server etcd-get failed")
		return nil, err
	} else {
		logger.GetLogger().Info("master server:init etcd success.")
		fmt.Println("master server init etcd success.")
	}

	server := &Server{
		Addr: fmt.Sprintf(":%d", defaultCfg.System.Addr),
	}

	server.setupSignal()

	//todo : set gin mode

	gin.SetMode(gin.DebugMode)

	return server, nil
}

func (srv *Server) RegisterRouters(routers ...func(engine *gin.Engine)) *Server {
	srv.Routers = append(srv.Routers, routers...)
	return srv
}

func (srv *Server) ListenAndServe() error {
	srv.Engine = gin.New()
	for _, c := range srv.Routers {
		c(srv.Engine)
	}

	srv.HttpServer = &http.Server{
		Handler:        srv.Engine,
		Addr:           srv.Addr,
		ReadTimeout:    20 * time.Second,
		WriteTimeout:   20 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	if err := srv.HttpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}
