package model

type PkgDecl struct {
	Id           int64  `xorm:"unique BIGINT(20)"`
	Pid          int64  `xorm:"unique(pkg_decl_pid_tag) unique(pkg_decl_pid) BIGINT(20)"`
	Tag          string `xorm:"unique(pkg_decl_pid_tag) VARCHAR(50)"`
	JsNum        int64  `xorm:"BIGINT(20)"`
	IsHasExport  int    `xorm:"TINYINT(1)"`
	IsHasConst   int    `xorm:"TINYINT(1)"`
	IsHasVar     int    `xorm:"TINYINT(1)"`
	IsHasExample int    `xorm:"TINYINT(1)"`
	Imports      string `xorm:"LONGTEXT"`
	TestImports  string `xorm:"LONGTEXT"`
	IsHasFile    int    `xorm:"TINYINT(1)"`
	IsHasSubdir  int    `xorm:"TINYINT(1)"`
}
