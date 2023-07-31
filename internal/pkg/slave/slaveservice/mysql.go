package slaveservice

import (
	"github.com/wyattjychen/hades/internal/pkg/model"
	"gorm.io/gorm"
)

func RegisterTables(db *gorm.DB) {
	_ = db.AutoMigrate(
		model.Node{},
		model.Job{},
		model.JobLog{},
		model.Script{},
	)
}
