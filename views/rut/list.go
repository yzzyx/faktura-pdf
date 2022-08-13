package rut

import (
	"github.com/yzzyx/faktura-pdf/models"
	"github.com/yzzyx/faktura-pdf/views"
)

// List is the view-handler for listing requests
type List struct {
	views.View
}

// NewList creates a new handler for listing requests
func NewList() *List {
	return &List{}
}

// HandleGet displays a list of requests
func (v *List) HandleGet() error {
	f := models.RUTFilter{}
	f.FilterStatus = []models.RUTStatus{
		models.RUTStatusPending,
		models.RUTStatusSent,
		models.RUTStatusRejected,
	}

	f.OrderBy = v.FormValueString("orderby")
	f.Direction = v.FormValueString("dir")
	f.CompanyID = v.Session.Company.ID
	filterPaid := false

	if v.FormValueBool("paid") {
		filterPaid = true
		f.FilterStatus = []models.RUTStatus{models.RUTStatusPaid}
	}

	rutRequests, err := models.RUTList(v.Ctx, f)
	if err != nil {
		return err
	}

	v.SetData("rutRequests", rutRequests)
	v.SetData("filterPaid", filterPaid)

	if v.FormValueBool("content") {
		return v.Render("rut/list-contents.html")
	}
	return v.Render("rut/list.html")
}
