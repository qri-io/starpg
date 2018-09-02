// Package skypg is a web-based skylark playground
package main

import (
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/skylark"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()

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

// ExecHandler assumes the request body is a skylark script to be executed
// currently no loader is provided, so all code must be defined inline
// errors are reported via HTTP response codes:
//   * 400: script errors
//   * 500: internal errors
func ExecHandler(w http.ResponseWriter, r *http.Request) {
	code, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error(err)
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	filename := filepath.Join(os.TempDir(), "exec.sky")
	if err := ioutil.WriteFile(filename, code, os.ModePerm); err != nil {
		log.Error(err.Error())
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	thread := &skylark.Thread{
		// print func writes directly to the response writer
		Print: func(thread *skylark.Thread, msg string) {
			w.Write([]byte(msg))
		},
	}

	if _, err = skylark.ExecFile(thread, filename, nil, nil); err != nil {
		msg := strings.Replace(err.Error(), filename, "line", 1)
		writeError(w, http.StatusBadRequest, errors.New(msg))
		return
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
	<title>Skylark Playground</title>
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
		#editor {
			flex: 1 1 50%;
			min-height: 400px;
			width: 100%;
		}
		#output {
			width: 100%;
			flex: 1 2 300px;
			padding: 25px 20px;
			overflow-y: auto;
			background: #f2f2f2;
			font-family: monospace;
		}
		.error { color: red; }
	</style>
</head>
<body>
	<div style="padding:10px">
		<button id="submit">Run</button>
		<h3>Skylark Playground</h3>
	</div>
	<div id="editor"></div>
	<div id="output"></div>

	<script src="/js/app.js"></script>
</body>
</html>`
