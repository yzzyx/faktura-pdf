package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/flosch/pongo2"
	"github.com/go-chi/chi/v5"
	"github.com/yzzyx/faktura-pdf/models"
)

func ViewInvoiceList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	f := models.InvoiceFilter{}
	f.OrderBy = r.FormValue("orderby")
	f.Direction = r.FormValue("dir")
	invoices, err := models.InvoiceListActive(ctx, f)
	if err != nil {
		RenderError(w, r, err)
		return
	}
	data := pongo2.Context{
		"invoices": invoices,
	}

	if r.FormValue("content") == "1" {
		Render("invoice-list-contents.html", w, r, data)
		return
	}
	Render("index.html", w, r, data)
}

func ViewInvoice(w http.ResponseWriter, r *http.Request) {
	var err error
	var invoice models.Invoice
	ctx := r.Context()
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	if id > 0 {
		invoice, err = models.InvoiceGet(ctx, id)
		if err != nil {
			RenderError(w, r, err)
			return
		}
	} else {
		invoice.Number, err = models.InvoiceGetNextNumber(ctx)
		if err != nil {
			RenderError(w, r, err)
			return
		}
	}

	data := pongo2.Context{
		"invoice":        invoice,
		"today":          time.Now(),
		"defaultDueDate": time.Now().AddDate(0, 1, 0),
	}

	if invoice.DateDue != nil {
		daysLeft := invoice.DateDue.Sub(time.Now()) / (time.Hour * 24)
		data["daysLeft"] = daysLeft
	}
	Render("invoice.html", w, r, data)
}

func ViewInvoiceOffer(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	invoice, err := models.InvoiceGet(ctx, id)
	if err != nil {
		RenderError(w, r, err)
		return
	}

	data, err := generatePDF(ctx, invoice, "offer.tex")
	if err != nil {
		RenderError(w, r, err)
		return
	}

	now := time.Now()
	name := fmt.Sprintf("offert-%d-%s-%s.pdf", invoice.Number, invoice.Name, now.Format("2006-01-02"))
	name = strings.ReplaceAll(name, " ", "_")
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename="+name)
	w.Write(data)
}

func ViewInvoiceInvoice(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	invoice, err := models.InvoiceGet(ctx, id)
	if err != nil {
		RenderError(w, r, err)
		return
	}

	data, err := generatePDF(ctx, invoice, "invoice.tex")
	if err != nil {
		RenderError(w, r, err)
		return
	}

	now := time.Now()
	name := fmt.Sprintf("faktura-%d-%s-%s.pdf", invoice.Number, invoice.Name, now.Format("2006-01-02"))
	name = strings.ReplaceAll(name, " ", "_")
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename="+name)
	w.Write(data)
}

func SetInvoiceFlag(w http.ResponseWriter, r *http.Request) {
	var err error
	var invoice models.Invoice

	ctx := r.Context()
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	invoice, err = models.InvoiceGet(ctx, id)
	if err != nil {
		RenderError(w, r, err)
		return
	}

	val := true

	if t, _ := strconv.ParseBool(r.FormValue("revoke")); t {
		val = false
	}

	flag := r.FormValue("flag")
	date := time.Now()
	if _, ok := r.Form["date"]; ok {
		if v, err := time.Parse("2006-01-02", r.FormValue("date")); err == nil {
			date = v
		}
	}

	switch flag {
	case "invoiced":
		invoice.IsInvoiced = val
		invoice.DateInvoiced = &date
	case "offered":
		invoice.IsOffered = val
	case "paid":
		invoice.IsPaid = val
		invoice.DatePaid = &date
	case "rut_sent":
		invoice.IsRutSent = val
	case "rut_paid":
		invoice.IsRutPaid = val
	default:
		RenderError(w, r, errors.New("invalid flag"))
		return
	}

	_, err = models.InvoiceSave(ctx, invoice)
	if err != nil {
		RenderError(w, r, err)
		return
	}

	w.Header().Set("Location", path.Join("/invoice", strconv.Itoa(id)))
	w.WriteHeader(http.StatusFound)
}

func SaveInvoice(w http.ResponseWriter, r *http.Request) {
	var err error
	var invoice models.Invoice

	ctx := r.Context()
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	updated := false

	err = r.ParseForm()
	if err != nil {
		RenderError(w, r, err)
		return
	}

	if id > 0 {
		invoice, err = models.InvoiceGet(ctx, id)
	} else {
		invoice.Number, err = models.InvoiceGetNextNumber(ctx)
		updated = true
	}
	if err != nil {
		RenderError(w, r, err)
		return
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
		_, ok := r.Form[formName]
		if !ok {
			continue
		}

		switch f := field.(type) {
		case *string:
			*f = r.FormValue(formName)
		case **time.Time:
			v := r.FormValue(formName)
			tv, err := time.Parse("2006-01-02", v)
			if err != nil {
				RenderError(w, r, err)
				return
			}
			*f = &tv
		default:
			err = fmt.Errorf("unknown field type %T, %v for field %s\n", f, f, formName)
			RenderError(w, r, err)
		}

		if !strings.HasPrefix(formName, "customer") {
			updated = true
		}
	}

	if v, _ := strconv.ParseBool(r.FormValue("rut_applicable_set")); v {
		invoice.RutApplicable, _ = strconv.ParseBool(r.FormValue("rut_applicable"))
		updated = true
	}

	invoice.Customer.ID, err = models.CustomerSave(ctx, invoice.Customer)
	if err != nil {
		RenderError(w, r, err)
		return
	}

	name := r.FormValue("name")
	if name != "" {
		invoice.Name = name
		updated = true
	}

	if updated {
		invoice.ID, err = models.InvoiceSave(ctx, invoice)
		if err != nil {
			RenderError(w, r, err)
			return
		}
	}

	// Check for new rows
	for _, rowData := range r.Form["row[]"] {
		var row models.InvoiceRow
		err := json.Unmarshal([]byte(rowData), &row)
		if err != nil {
			RenderError(w, r, err)
			return
		}

		err = models.InvoiceRowAdd(ctx, invoice.ID, row)
		if err != nil {
			RenderError(w, r, err)
			return
		}
	}

	// Check for deletes
	for _, str := range r.Form["delete_row[]"] {
		rowNumber, err := strconv.Atoi(str)
		if err != nil {
			RenderError(w, r, err)
			return
		}

		err = models.InvoiceRowRemove(ctx, invoice.ID, rowNumber)
		if err != nil {
			RenderError(w, r, err)
			return
		}
	}

	w.Header().Set("Location", fmt.Sprintf("/invoice/%d", invoice.ID))
	w.WriteHeader(http.StatusFound)
}
