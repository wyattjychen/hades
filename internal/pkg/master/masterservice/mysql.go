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
	// entities := []model.User{
	// 	{UserName: "root", Password: "e10adc3949ba59abbe56e057f20f883e", Role: model.RoleAdmin, Email: "333333333@qq.com"},
	// }
	// if exist := checkDataExist(db); !exist {
	// 	if err := db.Table(model.HadesUserTableName).Create(&entities).Error; err != nil {
	// 		return errors.Wrap(err, "Failed to initialize table data")
	// 	}
	// }
	logger.GetLogger().Info("register table success")
	return nil
}
