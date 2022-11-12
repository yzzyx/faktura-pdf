package invoice

import (
	"bytes"
	"fmt"
	"time"

	"github.com/yzzyx/faktura-pdf/models"
	"github.com/yzzyx/faktura-pdf/sie"
	"github.com/yzzyx/faktura-pdf/views"
)

// SIE is the view-handler for getting a SIE file for an invoice that can be imported to accounting systems
type SIE struct {
	views.View
}

// NewSIE creates a new handler for getting a SIE file
func NewSIE() *SIE {
	return &SIE{}
}

// HandleGet creates and sends a SIE file for the invoice that can be imported to accounting systems
func (v *SIE) HandleGet() error {
	var err error
	var invoice models.Invoice

	id := v.URLParamInt("id")

	invoice, err = models.InvoiceGet(v.Ctx, models.InvoiceFilter{ID: id, CompanyID: v.Session.Company.ID, IncludeCompany: true})
	if err != nil {
		return err
	}

	if !invoice.IsInvoiced {
		return fmt.Errorf("invoice is not marked as sent")
	}

	now := time.Now()
	if invoice.DateInvoiced == nil {
		invoice.DateInvoiced = &now
	}

	totals := invoice.Totals(true, true)

	export := sie.SIE{
		Flag:           0,
		Fnamn:          invoice.Company.Name,
		Program:        "FakturaPDF",
		ProgramVersion: "0.1",
		Type:           4,
		Verifications: []sie.Verification{
			{
				VerDatum: *invoice.DateInvoiced,
				VerText:  fmt.Sprintf("Faktura #%d - %s", invoice.Number, invoice.Name),
				Transactions: []sie.Transaction{
					{KontoNr: 1510, Belopp: totals.Customer},                                   // Kundfodringar
					{KontoNr: 1513, Belopp: totals.ROTRUT},                                     // Kundfodringar - delad faktura (ROT/RUT)
					{KontoNr: 2611, Belopp: totals.VAT25.Add(totals.ROTRUTTotals.VAT25).Neg()}, // Utgående moms på försäljning inom Sverige, 25 %
					{KontoNr: 2620, Belopp: totals.VAT12.Add(totals.ROTRUTTotals.VAT12).Neg()}, // Utgående moms 12 %
					{KontoNr: 2630, Belopp: totals.VAT6.Add(totals.ROTRUTTotals.VAT6).Neg()},   // Utgående moms 6 %
					{KontoNr: 3001, Belopp: totals.TotalVAT25.Neg()},                           // Försäljning varor inom Sverige, 25 % moms
					{KontoNr: 3002, Belopp: totals.TotalVAT12.Neg()},                           // Försäljning varor inom Sverige, 12 % moms
					{KontoNr: 3003, Belopp: totals.TotalVAT6.Neg()},                            // Försäljning varor inom Sverige, 6 % moms
					//{KontoNr: 3740, Belopp: 0},                 // Öres- och kronutjämning
				},
			},
		},
	}

	if invoice.IsPaid {
		if invoice.DatePaid == nil {
			invoice.DatePaid = &now
		}

		export.Verifications = append(export.Verifications, sie.Verification{
			VerDatum: *invoice.DatePaid,
			VerText:  fmt.Sprintf("Faktura #%d - %s betalad", invoice.Number, invoice.Name),
			Transactions: []sie.Transaction{
				{KontoNr: 1510, Belopp: totals.Customer.Neg()}, // Kundfodringar
				{KontoNr: 1930, Belopp: totals.Customer},       // Företags/affärskonto
			},
		})
	}

	b := &bytes.Buffer{}

	err = export.Write(b)
	if err != nil {
		return err
	}

	name := fmt.Sprintf("faktura-%d.si", invoice.Number)
	headers := v.ResponseHeaders()
	headers.Set("Content-Disposition", "attachment; filename="+name)
	return v.RenderBytes(b.Bytes())
}
