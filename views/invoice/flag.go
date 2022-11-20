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
	IsOffer bool
	views.View
}

// NewFlag creates a new handler for updating invoice flags
func NewFlag(isOffer bool) *Flag {
	return &Flag{IsOffer: isOffer}
}

// HandleGet updates the flags of an invoice
func (v *Flag) HandleGet() error {
	var err error
	var invoice models.Invoice

	id := v.URLParamInt("id")

	invoice, err = models.InvoiceGet(v.Ctx, models.InvoiceFilter{ID: id, CompanyID: v.Session.Company.ID, ListOffers: v.IsOffer})
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

	// Valid flags for invoices (0), offers (1), or both (2)
	validFlags := map[string]int{
		"invoiced": 0,
		"paid":     0,
		"offered":  1,
		"accepted": 1,
		"rejected": 1,
		"deleted":  2,
	}

	if valid, ok := validFlags[flag]; !ok || (valid == 0 && v.IsOffer) || (valid == 1 && !v.IsOffer) {
		return errors.New("invalid flag")
	}

	var createRUT bool
	switch flag {

	// Flags for invoices
	case "invoiced":
		invoice.IsInvoiced = val
		invoice.DateInvoiced = &date
	case "paid":
		invoice.IsPaid = val
		invoice.DatePaid = &date
		createRUT = invoice.RutApplicable && val

	// Flags for offers
	case "offered":
		invoice.Status = models.InvoiceStatusOffered
	case "accepted":
		invoice.Status = models.InvoiceStatusAccepted
	case "rejected":
		invoice.Status = models.InvoiceStatusRejected

	// Flags for both
	case "deleted":
		invoice.IsDeleted = val
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

	if invoice.IsDeleted {
		if v.IsOffer {
			return v.RedirectRoute("offer-list")
		}
		return v.RedirectRoute("invoice-list")
	}

	if v.IsOffer {
		return v.RedirectRoute("offer-view", "id", strconv.Itoa(id))
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
