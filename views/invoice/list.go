package invoice

import (
	"github.com/yzzyx/faktura-pdf/models"
	"github.com/yzzyx/faktura-pdf/views"
)

// List is the view-handler for listing invoices
type List struct {
	views.View
}

// NewList creates a new handler for listing invoices
func NewList() *List {
	return &List{}
}

// HandleGet displays a list of invoices
func (v *List) HandleGet() error {
	f := models.InvoiceFilter{}

	f.OrderBy = v.FormValueString("orderby")
	f.Direction = v.FormValueString("dir")
	f.CompanyID = v.Session.Company.ID

	filterPaid := false
	if v.FormValueBool("paid") {
		f.ListPaid = true
		filterPaid = true
	}

	invoices, err := models.InvoiceList(v.Ctx, f)
	if err != nil {
		return err
	}

	v.SetData("filterPaid", filterPaid)
	v.SetData("invoices", invoices)

	if v.FormValueBool("content") {
		return v.Render("invoice/list-contents.html")
	}
	return v.Render("invoice/list.html")
}
