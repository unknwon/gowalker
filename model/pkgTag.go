package model

type PkgTag struct {
	Id   int64  `xorm:"BIGINT(20)"`
	Path string `xorm:"unique(pkg_tag_path_tag) unique(pkg_tag_path) VARCHAR(150)"`
	Tag  string `xorm:"unique(pkg_tag_path_tag) VARCHAR(50)"`
	Vcs  string `xorm:"VARCHAR(50)"`
	Tags string `xorm:"LONGTEXT"`
}
