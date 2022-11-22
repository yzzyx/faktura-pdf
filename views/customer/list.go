package customer

import (
	"encoding/json"
	"strings"

	"github.com/yzzyx/faktura-pdf/models"
	"github.com/yzzyx/faktura-pdf/views"
	"github.com/yzzyx/zerr"
)

// List is the view-handler for listing companies
type List struct {
	views.View
}

// NewList creates a new handler for listing companies
func NewList() *List {
	return &List{}
}

// HandleGet displays an invoice
func (v *List) HandleGet() error {
	var err error
	var filter models.CustomerFilter

	filter.CompanyID = v.Session.Company.ID
	filter.Search = v.FormValueString("search")

	lst, err := models.CustomerList(v.Ctx, filter)
	if err != nil {
		return err
	}

	v.SetData("data", lst)
	v.SetData("redirect", v.FormValueString("r"))

	// FIXME - cleanup this code
	for _, a := range strings.Split(v.RequestHeaders().Get("accept"), ",") {
		if a == "application/json" {
			data, err := json.Marshal(lst)
			if err != nil {
				return zerr.Wrap(err)
			}
			return v.RenderBytes(data)
		}
	}

	if v.FormValueBool("content") {
		return v.Render("customer/list-contents.html")
	}
	return v.Render("customer/list.html")
}
