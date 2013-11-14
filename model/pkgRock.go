package model

type PkgRock struct {
	Id    int64  `xorm:"BIGINT(20)"`
	Pid   int64  `xorm:"unique(pkg_rock_pid) BIGINT(20)"`
	Path  string `xorm:"VARCHAR(150)"`
	Rank  int64  `xorm:"BIGINT(20)"`
	Delta int64  `xorm:"unique(pkg_rock_delta) BIGINT(20)"`
}
