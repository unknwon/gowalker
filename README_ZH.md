Go 步行者
========

Go 步行者是一个用于在线生成并浏览 <a target="_blank" href="http://golang.org/">Go</a> 项目 <b>源码</b> 文档的 Web 服务器，目前仅支持 Bitbucket、Github、Google Project Hosting 和 Launchpad 四大版本控制系统。

##主要功能
- 通过在首页的搜索框中输入外部包的 **导入路径** 或 **关键字** 进行搜索或在线生成文档。
- **在线生成** Go 项目文档：不需要附加安装任何组件即可开始使用。
- 当前包中公开函数和类型的 **鼠标悬浮提示** 和 **跳转链接**：减少寻找的时间。
- 对于当前包的公开函数和方法，拥有 **查看代码** 快速浏览功能：减少空间占用，让你可以在同时进行更多的工作。
- 对于外部包的公开函数和类型，拥有 **代码高亮** 和 **跳转链接** 的便捷功能：减少寻找的时间。
- 标准库导入路径 **输入提示**。
- **多语言** 文档支持（ **即将上线** ）。

##第三方包
- [Beego](https://github.com/astaxie/beego)：轻量级 Web 框架，适用于 Web 应用快速开发。
- [Qbs](https://github.com/coocood/qbs)：**结构化查询（Query by Struct）** 是一个非常给力的 ORM，目前支持 MySQL、SQLite3 和 PosgreSQL.
- [go-sqlite3](http://gowalker.org/github.com/mattn/go-sqlite3)：可用于 Go 开发的 SQLite3 数据库驱动；该驱动实现了 **`database/sql`** 接口，这对于代码迁移非常有利；该驱动采用 cgo 编写。

##多语言文档
即将上线

##授权许可
[Apache License, Version 2.0](http://www.apache.org/licenses/LICENSE-2.0.html)