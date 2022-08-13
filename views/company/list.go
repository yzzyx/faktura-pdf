package company

import (
	"github.com/yzzyx/faktura-pdf/models"
	"github.com/yzzyx/faktura-pdf/views"
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

	lst, err := models.CompanyList(v.Ctx, models.CompanyFilter{UserID: v.Session.User.ID})
	if err != nil {
		return err
	}
	v.SetData("companies", lst)
	v.SetData("redirect", v.FormValueString("r"))

	if v.FormValueBool("content") {
		return v.Render("company/list-contents.html")
	}
	return v.Render("company/list.html")
}
