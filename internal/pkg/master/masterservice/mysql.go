package masterservice

import (
	"fmt"
	"os"

	"github.com/wyattjychen/hades/internal/pkg/logger"
	"github.com/wyattjychen/hades/internal/pkg/model"
	"gorm.io/gorm"
)

func RegisterTables(db *gorm.DB) error {
	err := db.AutoMigrate(
		// model.User{},
		model.Node{},
		model.Job{},
		model.JobLog{},
		model.Script{},
	)
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("register table failed, error:%s", err.Error()))
		os.Exit(0)
	}

	logger.GetLogger().Info("register table success")
	return nil
}
