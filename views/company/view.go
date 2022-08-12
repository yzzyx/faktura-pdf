package company

import (
	"fmt"
	"strconv"
	"time"

	"github.com/yzzyx/faktura-pdf/models"
	"github.com/yzzyx/faktura-pdf/views"
)

// View is the view-handler for viewing an invoice
type View struct {
	views.View
}

// NewView creates a new handler for viewing an invoice
func NewView() *View {
	return &View{}
}

// HandleGet displays an invoice
func (v *View) HandleGet() error {
	var err error
	var company models.Company
	id := v.URLParamInt("id")

	// Set defaults
	company.InvoiceNumber = 1000
	company.PaymentType = models.PaymentTypeBG
	company.InvoiceDueDays = 30

	if id > 0 {
		company, err = models.CompanyGet(v.Ctx, id)
		if err != nil {
			return err
		}
	}

	v.SetData("c", company)

	// Used to create list of ROT/RUT services in invoice row modal
	v.SetData("rutServices", models.RUTServices)
	v.SetData("rotServices", models.ROTServices)
	v.SetData("defaultRUTService", models.RUTServiceTypeTradgardsarbete)
	v.SetData("defaultROTService", models.ROTServiceTypeBygg)

	return v.Render("company/view.html")
}

// HandlePost saves/updates an invoice
func (v *View) HandlePost() error {
	var err error
	var company models.Company

	id := v.URLParamInt("id")
	updated := false

	if id > 0 {
		company, err = models.CompanyGet(v.Ctx, id)
	}

	if err != nil {
		return err
	}

	fields := map[string]interface{}{
		"name":      &company.Name,
		"companyid": &company.CompanyID,
		"email":     &company.Email,
		"address1":  &company.Address1,
		"address2":  &company.Address2,
		"postcode":  &company.Postcode,
		"city":      &company.City,
		"telephone": &company.Telephone,

		"paymentaccount": &company.PaymentAccount,
		"paymenttype":    &company.PaymentType,
		"vatnumber":      &company.VATNumber,

		"invoicenumber":    &company.InvoiceNumber,
		"invoiceduedays":   &company.InvoiceDueDays,
		"invoicereference": &company.InvoiceReference,
		"invoicetext":      &company.InvoiceText,
	}

	for formName, field := range fields {
		if !v.FormValueExists(formName) {
			continue
		}

		switch f := field.(type) {
		case *int:
			*f = v.FormValueInt(formName)
		case *models.PaymentType:
			*f = models.PaymentType(v.FormValueInt(formName))
		case *string:
			*f = v.FormValueString(formName)
		case **time.Time:
			v := v.FormValueString(formName)
			tv, err := time.Parse("2006-01-02", v)
			if err != nil {
				return err
			}
			*f = &tv
		default:
			err = fmt.Errorf("unknown field type %T, %v for field %s\n", f, f, formName)
			return err
		}
		updated = true
	}

	isNewCompany := company.ID == 0

	if updated {
		company.ID, err = models.CompanySave(v.Ctx, company)
		if err != nil {
			return err
		}
	}

	if isNewCompany {
		err = company.AddUser(v.Ctx, v.Session.User)
		if err != nil {
			return err
		}

		v.Session.Company = company
		return v.RedirectRoute("start")
	}

	return v.RedirectRoute("company-view", "id", strconv.Itoa(company.ID))
}
