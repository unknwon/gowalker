package model

type PkgDoc struct {
	Id   int64  `xorm:"BIGINT(20)"`
	Path string `xorm:"unique(pkg_doc_path) VARCHAR(100)"`
	Lang string `xorm:"LONGTEXT"`
	Type string `xorm:"LONGTEXT"`
	Doc  string `xorm:"LONGTEXT"`
}
