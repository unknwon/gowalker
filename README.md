Go Walker
========
[中文文档](README_ZH.md)

Go Walker is a server that generates <a target="_blank" href="http://golang.org/">Go</a> projects <b>source code</b> documentation on the fly from projects on Bitbucket, Github, Google Project Hosting and Launchpad.

## Features

- Input the package **import path** or **keywords** in search boxes in home page to find or create documentation.
- Generate Go project documentation **on the fly** : no more installation required for using.
- **Mouse hover tip** and **jump link** for public functions and types in current package: reduce time to find.
- **Code view** for public functions, methods in current package in the same page: reduce rake up space and do more work at the same time.
- **Code highlight** and **jump link** for public functions, types from imported packages: reduce time to find.
- **Type-ahead** for standard library: not a big deal.

## Third-party packages

- [Beego](https://gowalker.org/github.com/astaxie/beego): a **lightweight** web framework for web application **quick** development.
- [Qbs](https://gowalker.org/github.com/coocood/qbs): **Query by Struct** is an excellent ORM, it supports MySQL, SQLite3 and PosgreSQL.
- [go-sqlite3](http://gowalker.org/github.com/mattn/go-sqlite3): SQLite3 database driver for Go; it **implemented `database/sql` interface** which is a very big deal for code migration; it requires cgo.

## Acknowledge

- [atotto](https://github.com/atotto) Thanks for translating site user interface to Japanese.

## Todo

- Add examples of functions.
- Add support for user submit examples.

## License

[Apache License, Version 2.0](http://www.apache.org/licenses/LICENSE-2.0.html), some part of code in this project is from [gopkgdoc](https://github.com/garyburd/gopkgdoc),see [File Change Log](FileChangeLog.md) for deatil.
