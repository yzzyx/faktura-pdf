package main

import (
	"net/http"
	"strconv"

	"github.com/flosch/pongo2"
	"github.com/go-chi/chi/v5"
	"github.com/yzzyx/faktura-pdf/models"
)

func ViewRutList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	f := models.RUTFilter{}
	f.FilterStatus = []models.RUTStatus{
		models.RUTStatusPending,
		models.RUTStatusSent,
		models.RUTStatusRejected,
	}

	f.OrderBy = r.FormValue("orderby")
	f.Direction = r.FormValue("dir")
	filterPaid := false

	if r.FormValue("paid") == "1" {
		filterPaid = true
		f.FilterStatus = []models.RUTStatus{models.RUTStatusPaid}
	}

	rutRequests, err := models.RUTList(ctx, f)
	if err != nil {
		RenderError(w, r, err)
		return
	}
	data := pongo2.Context{
		"rutRequests": rutRequests,
		"filterPaid":  filterPaid,
	}

	if r.FormValue("content") == "1" {
		Render("rut/list-contents.html", w, r, data)
		return
	}
	Render("rut/list.html", w, r, data)
}

func ViewRutRequest(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	rutRequest, err := models.RUTGet(ctx, id)
	if err != nil {
		RenderError(w, r, err)
		return
	}

	data := pongo2.Context{
		"rut": rutRequest,
	}

	Render("rut/view.html", w, r, data)
}
