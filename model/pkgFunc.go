package model

type PkgFunc struct {
	Id    int64  `xorm:"BIGINT(20)"`
	Pid   int64  `xorm:"unique(pkg_func_pid) BIGINT(20)"`
	Path  string `xorm:"VARCHAR(150)"`
	Name  string `xorm:"unique(pkg_func_name) VARCHAR(100)"`
	Doc   string `xorm:"LONGTEXT"`
	IsOld int    `xorm:"TINYINT(1)"`
}
