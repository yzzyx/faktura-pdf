package invoice

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
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
	var invoice models.Invoice
	id := v.URLParamInt("id")

	if id > 0 {
		invoice, err = models.InvoiceGet(v.Ctx, id)
		if err != nil {
			return err
		}
	} else {
		invoice.Number, err = models.InvoiceGetNextNumber(v.Ctx)
		if err != nil {
			return err
		}
	}

	v.SetData("invoice", invoice)
	v.SetData("today", time.Now())
	v.SetData("defaultDueDate", time.Now().AddDate(0, 1, 0))

	// Used to create list of ROT/RUT services in invoice row modal
	v.SetData("rutServices", models.RUTServices)
	v.SetData("rotServices", models.ROTServices)
	v.SetData("defaultRUTService", models.RUTServiceTypeTradgardsarbete)
	v.SetData("defaultROTService", models.ROTServiceTypeBygg)

	if invoice.DateDue != nil {
		daysLeft := invoice.DateDue.Sub(time.Now()) / (time.Hour * 24)
		v.SetData("daysLeft", daysLeft)
	}

	return v.Render("invoice/view.html")
}

// HandlePost saves/updates an invoice
func (v *View) HandlePost() error {
	var err error
	var invoice models.Invoice

	id := v.URLParamInt("id")
	updated := false

	if id > 0 {
		invoice, err = models.InvoiceGet(v.Ctx, id)
	} else {
		invoice.Number, err = models.InvoiceGetNextNumber(v.Ctx)
		updated = true
	}
	if err != nil {
		return err
	}

	fields := map[string]interface{}{
		"customer.name":     &invoice.Customer.Name,
		"customer.email":    &invoice.Customer.Email,
		"customer.address1": &invoice.Customer.Address1,
		"customer.address2": &invoice.Customer.Address2,
		"customer.postcode": &invoice.Customer.Postcode,
		"customer.city":     &invoice.Customer.City,
		"customer.pnr":      &invoice.Customer.PNR,
		"additional_info":   &invoice.AdditionalInfo,
		"date_due":          &invoice.DateDue,
		"date_invoiced":     &invoice.DateInvoiced,
	}

	for formName, field := range fields {
		if !v.FormValueExists(formName) {
			continue
		}

		switch f := field.(type) {
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

		if !strings.HasPrefix(formName, "customer") {
			updated = true
		}
	}

	if v.FormValueExists("rut_applicable_set") {
		invoice.RutApplicable = v.FormValueBool("rut_applicable_set")
		updated = true
	}

	invoice.Customer.ID, err = models.CustomerSave(v.Ctx, invoice.Customer)
	if err != nil {
		return err
	}

	name := v.FormValueString("name")
	if name != "" {
		invoice.Name = name
		updated = true
	}

	if updated {
		invoice.ID, err = models.InvoiceSave(v.Ctx, invoice)
		if err != nil {
			return err
		}
	}

	// Check for new rows
	for _, rowData := range v.FormValueStringSlice("row[]") {
		var row models.InvoiceRow
		err := json.Unmarshal([]byte(rowData), &row)
		if err != nil {
			return err
		}

		if row.ID > 0 {
			err = models.InvoiceRowUpdate(v.Ctx, row)
		} else {
			err = models.InvoiceRowAdd(v.Ctx, invoice.ID, row)
		}
		if err != nil {
			return err
		}
	}

	// Check for deletes
	for _, str := range v.FormValueStringSlice("delete_row[]") {
		rowNumber, err := strconv.Atoi(str)
		if err != nil {
			return err
		}

		err = models.InvoiceRowRemove(v.Ctx, invoice.ID, rowNumber)
		if err != nil {
			return err
		}
	}

	return v.RedirectRoute("invoice-view", "id", strconv.Itoa(invoice.ID))
}
