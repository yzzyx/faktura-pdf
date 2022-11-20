package invoice

import (
	"github.com/yzzyx/faktura-pdf/models"
	"github.com/yzzyx/faktura-pdf/views"
)

// List is the view-handler for listing invoices
type List struct {
	IsOffer bool
	views.View
}

// NewList creates a new handler for listing invoices
func NewList(isOffer bool) *List {
	return &List{IsOffer: isOffer}
}

// HandleGet displays a list of invoices
func (v *List) HandleGet() error {
	f := models.InvoiceFilter{}

	f.OrderBy = v.FormValueString("orderby")
	f.Direction = v.FormValueString("dir")
	f.CompanyID = v.Session.Company.ID
	f.ListOffers = v.IsOffer

	filterPaid := false
	f.FilterPaid = 2
	if v.FormValueBool("paid") {
		f.FilterPaid = 1
		filterPaid = true
	}

	invoices, err := models.InvoiceList(v.Ctx, f)
	if err != nil {
		return err
	}

	v.SetData("filterPaid", filterPaid)
	v.SetData("invoices", invoices)
	v.SetData("isOffer", v.IsOffer)

	if v.FormValueBool("content") {
		return v.Render("invoice/list-contents.html")
	}
	return v.Render("invoice/list.html")
}
