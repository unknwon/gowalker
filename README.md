Go Walker
========
[中文文档](README_ZH.md)

Go Walker is a web server that generates <a target="_blank" href="http://golang.org/">Go</a> projects API documentation with source code on the fly from projects on <b><a target="_blank" href="https://bitbucket.org/">Bitbucket</a></b>, <b><a target="_blank" href="https://github.com/">Github</a></b>, <b><a target="_blank" href="http://code.google.com/">Google Code</a></b>, <b><a target="_blank" href="https://launchpad.net/">Launchpad</a></b>, <b><a target="_blank" href="http://git.oschina.net/">Git @ OSC</a></b> and <b><a target="_blank" href="https://code.csdn.net/">CSDN Code</a></b>.

![](https://github.com/Unknwon/gowalker/blob/master/docs/images/whatisthis.png?raw=true)

## Features

- Input the package **import path** or **keywords** in search boxes in home page to find or create documentation.
- Generate Go project documentation **on the fly** : no more installation required for using.
- **Mouse hover tip** and **jump link** for public functions and types in current package: reduce time to find.
- **Code view** for public functions, methods in current package in the same page: reduce rake up space and do more work at the same time.
- **Code highlight** and **jump link** for public functions, types from imported packages: reduce time to find.
- **Type-ahead** for standard library: not a big deal.
- **Control panel** for keyboard shortcuts: compatible with godoc.org.
- Use **Tag** to label your project: list by categories.

## Third-party packages

- [Beego](http://gowalker.org/github.com/astaxie/beego): a **lightweight** web framework for web application **quick** development.
- [Qbs](http://gowalker.org/github.com/coocood/qbs): **Query by Struct** is an excellent ORM, it supports MySQL, SQLite3 and PosgreSQL.
- [go-sqlite3](http://gowalker.org/github.com/mattn/go-sqlite3): SQLite3 database driver for Go; it **implemented `database/sql` interface** which is a very big deal for code migration; it requires cgo.

## Credits

- [chenwenli](http://www.lavachen.cn) Thanks for adding feature of [Control Panel](http://gowalker.org/about#control_panel).
- [atotto](https://github.com/atotto) Thanks for translating site user interface to Japanese.
- Source files that contain code that is from [gddo](https://github.com/garyburd/gddo) is honored in specific.

## Todo

- Add specialized markdown to HTML function for parsing readme in gowalker.

## License

[Apache License, Version 2.0](http://www.apache.org/licenses/LICENSE-2.0.html).
