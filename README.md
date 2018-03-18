mkdeb [![][godoc-badge]][godoc] [![][report-badge]][report]
=====

[mkdeb][project] is an open source application to generate Debian packages from upstream release archives.

The source code is available at [Github][source] and is licensed under the terms of the [Apache License 2.0][license].

Installation
------------

To install mkdeb, run:

    go get mkdeb.sh/cmd/mkdeb

Recipes
-------

Packages are built using recipes provided in a separate Git [repository][recipes].


[godoc-badge]:  https://godoc.org/github.com/mkdeb/mkdeb?status.svg
[godoc]:        https://godoc.org/github.com/mkdeb/mkdeb
[license]:      https://www.apache.org/licenses/LICENSE-2.0
[project]:      https://mkdeb.sh/
[recipes]:      https://github.com/mkdeb/recipes
[report-badge]: https://goreportcard.com/badge/github.com/mkdeb/mkdeb
[report]:       https://goreportcard.com/report/github.com/mkdeb/mkdeb
[source]:       https://github.com/mkdeb/mkdeb
