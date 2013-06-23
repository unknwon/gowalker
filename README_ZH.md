Go 步行者
========

Go 步行者是一个用于在线生成并浏览 <a target="_blank" href="http://golang.org/">Go</a> 项目 API 文档及源码 的 Web 服务器，目前支持包括 <b><a target="_blank" href="https://bitbucket.org/">Bitbucket</a></b>，<b><a target="_blank" href="https://github.com/">Github</a></b>，<b><a target="_blank" href="http://code.google.com/">Google Code</a></b>，<b><a target="_blank" href="https://launchpad.net/">Launchpad</a></b>，<b><a target="_blank" href="http://git.oschina.net/">Git @ OSC</a></b> 和 <b><a target="_blank" href="https://code.csdn.net/">CSDN Code</a></b> 六大版本控制平台。

![](https://github.com/Unknwon/gowalker/blob/master/docs/images/whatisthis_ZH.png?raw=true)

## 主要特性

- 在超过 7000 个海量项目中 **搜索** 并 **查看** 文档。
- 通过在首页的搜索框中输入外部包的 **导入路径** 或 **关键字** 进行搜索或在线生成文档。
- **在线生成** Go 项目文档：不需要附加安装任何组件即可开始使用。
- 当前包中公开函数和类型的 **鼠标悬浮提示** 和 **跳转链接**：减少寻找的时间。
- 对于当前包的公开函数和方法，拥有 **查看代码** 快速浏览功能：减少空间占用，让你可以在同时进行更多的工作。
- 对于外部包的公开函数和类型，拥有 **代码高亮** 和 **跳转链接** 的便捷功能：减少寻找的时间。
- 标准库导入路径 **输入提示**。
- 键盘快捷键 **控制面板**：与 godoc.org 兼容。
- 使用 **标签** 标记项目：分类显示。
- [完整文档](https://github.com/Unknwon/gowalker/blob/master/docs/Features_ZH.md).


## 第三方包

- [Beego](http://gowalker.org/github.com/astaxie/beego)：轻量级 Web 框架，适用于 Web 应用快速开发。
- [Qbs](http://gowalker.org/github.com/coocood/qbs)： **结构化查询（Query by Struct）** 是一个非常给力的 ORM，目前支持 MySQL、SQLite3 和 PosgreSQL.
- [go-sqlite3](http://gowalker.org/github.com/mattn/go-sqlite3)：可用于 Go 开发的 SQLite3 数据库驱动；该驱动实现了 **`database/sql`** 接口，这对于代码迁移非常有利；该驱动采用 cgo 编写。

## 授权许可

[Apache License, Version 2.0](http://www.apache.org/licenses/LICENSE-2.0.html)
