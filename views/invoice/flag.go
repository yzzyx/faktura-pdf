package invoice

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/yzzyx/faktura-pdf/models"
	"github.com/yzzyx/faktura-pdf/views"
)

// Flag is the view-handler for updating invoice flags
type Flag struct {
	views.View
}

// NewFlag creates a new handler for updating invoice flags
func NewFlag() *Flag {
	return &Flag{}
}

// HandleGet updates the flags of an invoice
func (v *Flag) HandleGet() error {
	var err error
	var invoice models.Invoice

	id := v.URLParamInt("id")

	invoice, err = models.InvoiceGet(v.Ctx, models.InvoiceFilter{ID: id, CompanyID: v.Session.Company.ID})
	if err != nil {
		return err
	}

	val := true

	if v.FormValueBool("revoke") {
		val = false
	}

	flag := v.FormValueString("flag")
	date := time.Now()
	if v.FormValueExists("date") {
		if v, err := time.Parse("2006-01-02", v.FormValueString("date")); err == nil {
			date = v
		}
	}

	var createRUT bool
	switch flag {
	case "invoiced":
		invoice.IsInvoiced = val
		invoice.DateInvoiced = &date
	case "offered":
		invoice.IsOffered = val
	case "paid":
		invoice.IsPaid = val
		invoice.DatePaid = &date
		createRUT = invoice.RutApplicable && val
	default:
		return errors.New("invalid flag")
	}

	_, err = models.InvoiceSave(v.Ctx, invoice)
	if err != nil {
		return err
	}

	if createRUT {
		err = createROTRUTFromInvoice(v.Ctx, invoice)
		if err != nil {
			return err
		}
	}

	return v.RedirectRoute("invoice-view", "id", strconv.Itoa(id))
}

// HandlePost updates the flags of an invoice
func (v *Flag) HandlePost() error {
	return v.HandleGet()
}

func createROTRUTFromInvoice(ctx context.Context, invoice models.Invoice) error {
	typeRows := map[models.RUTType][]models.InvoiceRow{}

	for _, r := range invoice.Rows {
		if !r.IsRotRut || r.RotRutServiceType == nil {
			continue
		}

		if r.RotRutServiceType.IsRUT() {
			typeRows[models.RUTTypeRUT] = append(typeRows[models.RUTTypeRUT], r)
		} else {
			typeRows[models.RUTTypeROT] = append(typeRows[models.RUTTypeROT], r)
		}
	}

	for typ, _ := range typeRows {
		lst, err := models.RUTList(ctx, models.RUTFilter{InvoiceID: invoice.ID, Type: &typ})
		if err != nil {
			return err
		}

		// Only create a RUT-request if we don't already have one
		if len(lst) == 0 {
			rut := models.RUT{
				Invoice: invoice,
				Type:    typ,
			}

			_, err := models.RUTSave(ctx, rut)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
