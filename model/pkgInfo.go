package model

import (
	"time"
)

type PkgInfo struct {
	Id          int64     `xorm:"BIGINT(20)"`
	ImportPath  string    `xorm:"unique(pkg_info_import_path) VARCHAR(150)"`
	ProjectName string    `xorm:"VARCHAR(50)"`
	ProjectPath string    `xorm:"VARCHAR(120)"`
	ViewDirPath string    `xorm:"VARCHAR(120)"`
	Synopsis    string    `xorm:"VARCHAR(300)"`
	IsCmd       int       `xorm:"TINYINT(1)"`
	IsGoRepo    int       `xorm:"TINYINT(1)"`
	IsGoSubrepo int       `xorm:"TINYINT(1)"`
	Views       int64     `xorm:"unique(pkg_info_views) BIGINT(20)"`
	ViewedTime  int64     `xorm:"BIGINT(20)"`
	Created     time.Time `xorm:"not null default CURRENT_TIMESTAMP unique(pkg_info_created) TIMESTAMP"`
	Rank        int64     `xorm:"unique(pkg_info_rank) BIGINT(20)"`
	PkgVer      int64     `xorm:"BIGINT(20)"`
	Ptag        string    `xorm:"VARCHAR(50)"`
	Labels      string    `xorm:"LONGTEXT"`
	RefNum      int64     `xorm:"BIGINT(20)"`
	RefPids     string    `xorm:"LONGTEXT"`
	Homepage    string    `xorm:"VARCHAR(100)"`
	ForkUrl     string    `xorm:"VARCHAR(150)"`
	Issues      int64     `xorm:"BIGINT(20)"`
	Stars       int64     `xorm:"BIGINT(20)"`
	Forks       int64     `xorm:"BIGINT(20)"`
	Note        string    `xorm:"LONGTEXT"`
	SourceSize  int64     `xorm:"BIGINT(20)"`
}
