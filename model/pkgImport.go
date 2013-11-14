package model

type PkgImport struct {
	Id      int64  `xorm:"BIGINT(20)"`
	Path    string `xorm:"unique(pkg_import_path) VARCHAR(150)"`
	Imports string `xorm:"LONGTEXT"`
}
