package model

import (
	"fmt"
	"strings"

	"github.com/wyattjychen/hades/internal/pkg/mysqlconn"
	"github.com/wyattjychen/hades/internal/pkg/utils"
	"github.com/wyattjychen/hades/internal/pkg/utils/hadeserrors"
)

// Preset Script
type Script struct {
	ID      int    `json:"id" gorm:"column:id;primary_key;auto_increment"`
	Name    string `json:"name" gorm:"size:256;column:name;not null;index:idx_script_name" binding:"required"`
	Command string `json:"command" gorm:"type:text;column:command;not null" binding:"required"`
	Created int64  `json:"created" gorm:"column:created;not null"`
	Updated int64  `json:"updated" gorm:"column:updated;default:0"`

	Cmd []string `json:"cmd" gorm:"-"`
}

func (s *Script) Insert() (insertId int, err error) {
	err = mysqlconn.GetMysqlDB().Table(HadesScriptTableName).Create(s).Error
	if err == nil {
		insertId = s.ID
	}
	return
}

func (s *Script) Update() error {
	return mysqlconn.GetMysqlDB().Table(HadesScriptTableName).Updates(s).Error
}

func (s *Script) Delete() error {
	return mysqlconn.GetMysqlDB().Exec(fmt.Sprintf("delete from %s where id = ?", HadesScriptTableName), s.ID).Error
}

func (s *Script) FindById() error {
	return mysqlconn.GetMysqlDB().Table(HadesScriptTableName).Where("id = ? ", s.ID).First(s).Error
}

func (s *Script) TableName() string {
	return HadesScriptTableName
}

func (s *Script) Check() error {
	s.Name = strings.TrimSpace(s.Name)
	if len(s.Name) == 0 {
		return hadeserrors.EmptyScriptNameErr
	}
	if len(strings.TrimSpace(s.Command)) == 0 {
		return hadeserrors.EmptyScriptCommandErr
	}
	if len(s.Cmd) == 0 {
		s.SplitCmd()
	}
	return nil
}

func (s *Script) SplitCmd() {
	ps := strings.SplitN(s.Command, " ", 2)
	if len(ps) == 1 {
		s.Cmd = ps
		return
	}
	s.Cmd = make([]string, 0, 2)
	s.Cmd = append(s.Cmd, ps[0])
	s.Cmd = append(s.Cmd, utils.ParseCmdArguments(ps[1])...)
}
