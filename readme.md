# Starlark Playground

Starlark Playground is a web-based starlark editor. It uses the golang implementation of starlark running on a server to present a  [monaco](https://github.com/Microsoft/monaco-editor) editor set to `python` syntax.

### Getting Started

You'll need a recent version of [yarn](https://yarnpkg.com) and [go](https://golang.org). With those installed, run:

```shell
$ go get github.com/qri-io/starpg
$ cd $GOPATH/github.com/qri-io/starpg
$ make
```

You'll see _lots_ of output as the makefile uses yarn to grab bunch of dependencies and build the frontend editor. It'll then install any missing go dependencies and spin up a server. If you see this you're in business:

```shell
INFO[0000] starting editor on port 3000
```
From there use a browser to visit `http://localhost:3000`. Happy starlark editing!
