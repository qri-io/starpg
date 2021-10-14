// Package skypg is a web-based starlark playground
package main

import (
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/qri-io/starlib"
	"github.com/sirupsen/logrus"
	"go.starlark.net/resolve"
	"go.starlark.net/starlark"
)

var log = logrus.New()

func init() {
	// hoist execution settings to resolve package settings
	resolve.AllowSet = true
	resolve.AllowGlobalReassign = true
	resolve.AllowRecursion = true
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Infof("starting editor on port %s", port)
	err := http.ListenAndServe(":"+port, NewMux())
	if err != nil {
		log.Error(err.Error())
	}
}

// NewMux creates a muxer with this server's designated routes
func NewMux() *http.ServeMux {
	m := http.NewServeMux()
	m.Handle("/", LogRequest(HomeHandler))
	m.Handle("/exec", LogRequest(ExecHandler))
	m.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("editor/dist"))))

	return m
}

// LogRequest is a middleware func for writing requests via the logger
func LogRequest(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Infof("%s\t%s", r.Method, r.URL.Path)
		f(w, r)
	}
}

// HomeHandler writes the response HTML template
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(tmpl))
}

// ExecHandler assumes the request body is a starlark script to be executed
// currently no loader is provided, so all code must be defined inline
// errors are reported via HTTP response codes:
//   * 400: script errors
//   * 500: internal errors
func ExecHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	f, err := ioutil.TempFile("", "exec_starlark")
	if err != nil {
		log.Error(err.Error())
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	defer os.Remove(f.Name())
	if _, err := io.Copy(f, r.Body); err != nil {
		log.Error(err.Error())
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	wrote := false
	thread := &starlark.Thread{
		// print func writes directly to the response writer
		Print: func(thread *starlark.Thread, msg string) {
			w.Write([]byte(msg))
			wrote = true
		},
		Load: starlib.Loader,
	}

	if _, err = starlark.ExecFile(thread, f.Name(), nil, nil); err != nil {
		msg := strings.Replace(err.Error(), f.Name(), "line", 1)
		writeError(w, http.StatusBadRequest, errors.New(msg))
		return
	}

	if wrote == false {
		w.Write([]byte("(no output)"))
	}
}

// writeError writes a status code
func writeError(w http.ResponseWriter, status int, err error) {
	w.WriteHeader(status)
	w.Write([]byte(err.Error()))
}

// tmpl is the home template, inlined so we have one less file to deal with
const tmpl = `<!DOCTYPE html>
<html>
<head>
	<meta charset="utf-8" />
	<meta http-equiv="X-UA-Compatible" content="IE=edge">
	<title>Starlark Playground</title>
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<style>
		* {
			box-sizing: border-box;
		}
		html, body {
			height: 100%;
			margin: 0; 
			padding: 0;
			font-family: "avenir-next", helvetica, sans-serif;
			display: flex;
			flex-direction: column;
		}
		#topbar {
			padding:10px;
			box-shadow: 0 0 2px rgba(0,0,0,0.2);
			z-index: 10;
		}
		#submit {
			float: right;
			vertical-align: top;
			padding: 10px 20px;
			border-radius: 3px;
			font-weight: bold;
			font-size: 16px;
			color: white;
			text-align: center;
			background: #2980b9;
			border: 0;
			border-bottom: 2px solid #2475ab;
			cursor: pointer;
			box-shadow: inset 0 -2px #2475ab;
		}
		#panes {
			flex: 3 3 80%;
			width: 100%;
			overflow: hidden;
			display: flex;
		}
		#editor {
			flex: 1 1 60%;
			min-height: 400px;
		}
		#output {
			flex: 1 2 40%;
			padding: 25px 20px;
			overflow-y: auto;
			background: #f2f2f2;
			font-family: monospace;
		}
		.error { color: red; }
	</style>
</head>
<body>
	<div id="topbar">
		<button id="submit">Run</button>
		<h3>Starlark Playground</h3>
	</div>
	<div id="panes">
		<div id="editor"></div>
		<div id="output"></div>
	</div>

	<script src="/js/app.js"></script>
</body>
</html>`
