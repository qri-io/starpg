# Skylark Playground

Skylark Playground is a web-based skylark editor. It uses the golang implementation of skylark running on a server to present a  [monaco](https://github.com/Microsoft/monaco-editor) editor set to `python` syntax.

### Getting Started

You'll need a recent version of [yarn](https://yarnpkg.com) and [go](https://golang.org). With those installed, run:

```shell
$ go get github.com/qri-io/skypg
$ cd $GOPATH/github.com/qri-io/skypg
$ make
```

You'll see _lots_ of output as the makefile uses yarn to grab bunch of dependencies and build the frontend editor. It'll then install any missing go dependencies and spin up a server. If you see this you're in business:

```shell
INFO[0000] starting editor on port 3000
```
From there use a browser to visit `http://localhost:3000`. Happy skylark editing!
