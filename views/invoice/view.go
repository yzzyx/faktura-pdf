package invoice

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/yzzyx/faktura-pdf/models"
	"github.com/yzzyx/faktura-pdf/views"
)

// View is the view-handler for viewing an invoice
type View struct {
	IsOffer bool
	views.View
}

// NewView creates a new handler for viewing an invoice
func NewView(isOffer bool) *View {
	return &View{IsOffer: isOffer}
}

// HandleGet displays an invoice
func (v *View) HandleGet() error {
	var err error
	var invoice models.Invoice
	id := v.URLParamInt("id")

	if id > 0 {
		invoice, err = models.InvoiceGet(v.Ctx, models.InvoiceFilter{ID: id, CompanyID: v.Session.Company.ID, ListOffers: v.IsOffer})
		if err != nil {
			return err
		}
	} else {
		invoice.Number, err = v.Session.Company.GetNextInvoiceNumber(v.Ctx)
		if err != nil {
			return err
		}
	}

	v.SetData("invoice", invoice)
	v.SetData("totals", invoice.Totals(false, false))
	v.SetData("today", time.Now())
	v.SetData("defaultDueDate", time.Now().AddDate(0, 1, 0))
	v.SetData("isOffer", v.IsOffer)

	if invoice.ID > 0 {
		attachments, err := models.FileList(v.Ctx, models.FileFilter{
			CompanyID:      v.Session.Company.ID,
			InvoiceID:      invoice.ID,
			IncludeContent: false,
		})
		if err != nil {
			return err
		}
		v.SetData("attachments", attachments)

	}

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
		invoice, err = models.InvoiceGet(v.Ctx, models.InvoiceFilter{ID: id, CompanyID: v.Session.Company.ID, ListOffers: v.IsOffer})
	} else {
		invoice.Number, err = v.Session.Company.GetNextInvoiceNumber(v.Ctx)
		invoice.Company.ID = v.Session.Company.ID
		invoice.Customer.CompanyID = v.Session.Company.ID
		invoice.IsOffer = v.IsOffer
		updated = true
	}
	if err != nil {
		return err
	}

	var customerID int
	fields := map[string]interface{}{
		"customer.id":        &customerID,
		"customer.name":      &invoice.Customer.Name,
		"customer.email":     &invoice.Customer.Email,
		"customer.address1":  &invoice.Customer.Address1,
		"customer.address2":  &invoice.Customer.Address2,
		"customer.postcode":  &invoice.Customer.Postcode,
		"customer.city":      &invoice.Customer.City,
		"customer.pnr":       &invoice.Customer.PNR,
		"customer.telephone": &invoice.Customer.Telephone,
		"additional_info":    &invoice.AdditionalInfo,
		"date_due":           &invoice.DateDue,
		"date_invoiced":      &invoice.DateInvoiced,
	}

	customerUpdated := false
	for formName, field := range fields {
		if !v.FormValueExists(formName) {
			continue
		}

		switch f := field.(type) {
		case *int:
			*f = v.FormValueInt(formName)
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
		} else if formName != "customer.id" {
			customerUpdated = true
		}
	}

	if v.FormValueExists("rut_applicable_set") {
		invoice.RutApplicable = v.FormValueBool("rut_applicable")
		updated = true
	}

	// Validate customer change
	if customerID != invoice.Customer.ID {
		if customerID > 0 {
			c, err := models.CustomerList(v.Ctx, models.CustomerFilter{
				ID:        customerID,
				CompanyID: v.Session.Company.ID,
			})
			if err != nil {
				return err
			}

			if len(c) != 1 {
				return fmt.Errorf("invalid customer selection")
			}

			invoice.Customer.ID = customerID
			updated = true
		} else if customerUpdated {
			invoice.Customer.ID = 0
			updated = true
		}

	}

	if customerUpdated {
		invoice.Customer.ID, err = models.CustomerSave(v.Ctx, invoice.Customer)
		if err != nil {
			return err
		}
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

	// Only allow updates to invoices we haven't sent yet
	if !invoice.IsInvoiced {
		err = v.updateRows(invoice)
		if err != nil {
			return err
		}
	}

	for _, upload := range v.FormFiles("attachment") {
		f, err := upload.Open()
		if err != nil {
			return err
		}

		mimeType := mime.TypeByExtension(filepath.Ext(upload.Filename))
		attachment := models.File{
			Name:      upload.Filename,
			CompanyID: v.Session.Company.ID,
			MIMEType:  mimeType,
			Backend:   nil,
		}

		attachment.Contents, err = ioutil.ReadAll(f)
		if err != nil {
			return err
		}

		err = models.InvoiceAddAttachment(v.Ctx, invoice, attachment)
		if err != nil {
			return err
		}
	}

	// Check for any removed attachments
	for _, remove := range v.FormValueStringSlice("removeAttachment[]") {
		id, err := strconv.Atoi(remove)
		if err != nil {
			continue
		}

		err = models.InvoiceRemoveAttachment(v.Ctx, invoice, id)
		if err != nil {
			return err
		}
	}

	if v.IsOffer {
		return v.RedirectRoute("offer-view", "id", strconv.Itoa(invoice.ID))
	}

	return v.RedirectRoute("invoice-view", "id", strconv.Itoa(invoice.ID))
}

func (v *View) updateRows(invoice models.Invoice) error {
	var err error

	updatedRows := map[int]*models.InvoiceRow{}
	rowOrderText := v.FormValueString("roworder")
	if rowOrderText != "" {
		rowOrder := []int{}
		err = json.Unmarshal([]byte(rowOrderText), &rowOrder)
		if err != nil {
			return err
		}

		for k, id := range rowOrder {
			for rowIdx := range invoice.Rows {
				r := &invoice.Rows[rowIdx]
				if r.ID != id || r.RowOrder == k {
					continue
				}

				r.RowOrder = k
				updatedRows[r.ID] = r
			}
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
			// This row has also been moved
			if v, ok := updatedRows[row.ID]; ok {
				row.RowOrder = v.RowOrder
			}
			updatedRows[row.ID] = &row
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

		delete(updatedRows, rowNumber)
		err = models.InvoiceRowRemove(v.Ctx, invoice.ID, rowNumber)
		if err != nil {
			return err
		}
	}

	for _, row := range updatedRows {
		err = models.InvoiceRowUpdate(v.Ctx, *row)
		if err != nil {
			return err
		}
	}

	return nil
}
