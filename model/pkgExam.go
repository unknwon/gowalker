package model

import (
	"time"
)

type PkgExam struct {
	Id       int64     `xorm:"BIGINT(20)"`
	Path     string    `xorm:"unique(pkg_exam_path) VARCHAR(150)"`
	Gist     string    `xorm:"VARCHAR(150)"`
	Examples string    `xorm:"LONGTEXT"`
	Created  time.Time `xorm:"not null default CURRENT_TIMESTAMP unique(pkg_exam_created) TIMESTAMP"`
}
