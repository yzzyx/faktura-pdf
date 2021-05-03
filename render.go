package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/flosch/pongo2"
)

var templateSet *pongo2.TemplateSet

func TemplateSetup() {
	fsLoader, err := pongo2.NewLocalFileSystemLoader("templates")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %+v\n", err)
		return
	}

	templateSet = pongo2.NewSet("base", fsLoader)
}

func Render(file string, w http.ResponseWriter, r *http.Request, data pongo2.Context) {
	tmpl, err := templateSet.FromFile(file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.ExecuteWriter(data, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
func RenderError(w http.ResponseWriter, r *http.Request, err error) {
	fmt.Fprintf(os.Stderr, "[%s] error: %v\n", r.RemoteAddr, err)
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
