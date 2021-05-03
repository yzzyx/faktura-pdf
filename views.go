package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/flosch/pongo2"
	"github.com/go-chi/chi/v5"
	"github.com/shopspring/decimal"
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

	updated := false
	if r.Method == "POST" {
		invoiceDateStr := r.FormValue("invoiceDate")
		dueDateStr := r.FormValue("dueDate")
		preview, _ := strconv.ParseBool(r.FormValue("preview"))

		if invoiceDateStr != "" {
			v, err := time.Parse("2006-01-02", invoiceDateStr)
			if err != nil {
				RenderError(w, r, err)
				return
			}
			invoice.DateInvoiced = &v
			updated = true
		}

		if dueDateStr != "" {
			v, err := time.Parse("2006-01-02", dueDateStr)
			if err != nil {
				RenderError(w, r, err)
				return
			}
			invoice.DateDue = &v
			updated = true
		}

		if preview {
			updated = false
		}
	}

	data, err := generatePDF(ctx, invoice, "invoice.tex")
	if err != nil {
		RenderError(w, r, err)
		return
	}

	if updated {
		invoice.IsInvoiced = true
		_, err = models.InvoiceSave(ctx, invoice)
		if err != nil {
			RenderError(w, r, err)
			return
		}
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
	now := time.Now()
	switch flag {
	case "invoiced":
		invoice.IsInvoiced = val
		invoice.DateInvoiced = &now
	case "offered":
		invoice.IsOffered = val
	case "payed":
		invoice.IsPayed = val
		invoice.DatePayed = &now
	default:
		RenderError(w, r, errors.New("invalid flag"))
		return
	}

	_, err = models.InvoiceSave(ctx, invoice)
	if err != nil {
		RenderError(w, r, err)
		return
	}

	w.Header().Set("Location", "/")
	w.WriteHeader(http.StatusFound)
}

func SaveInvoice(w http.ResponseWriter, r *http.Request) {
	var err error
	var invoice models.Invoice

	ctx := r.Context()
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	err = r.ParseForm()
	if err != nil {
		RenderError(w, r, err)
		return
	}

	if id > 0 {
		invoice, err = models.InvoiceGet(ctx, id)
	} else {
		invoice.Number, err = models.InvoiceGetNextNumber(ctx)
	}
	if err != nil {
		RenderError(w, r, err)
		return
	}

	fields := map[string]*string{
		"customer.name":     &invoice.Customer.Name,
		"customer.email":    &invoice.Customer.Email,
		"customer.address1": &invoice.Customer.Address1,
		"customer.address2": &invoice.Customer.Address2,
		"customer.postcode": &invoice.Customer.Postcode,
		"customer.city":     &invoice.Customer.City,
		"additional_info":   &invoice.AdditionalInfo,
	}
	for formName, field := range fields {
		_, ok := r.Form[formName]
		if !ok {
			continue
		}
		*field = r.FormValue(formName)
	}

	invoice.Customer.ID, err = models.CustomerSave(ctx, invoice.Customer)
	if err != nil {
		RenderError(w, r, err)
		return
	}

	name := r.FormValue("name")
	if name != "" {
		invoice.Name = name
	}

	// FIXME - cleanup "invoice-changed"-check
	if name != "" || r.FormValue("additional_info") != "" {
		invoice.ID, err = models.InvoiceSave(ctx, invoice)
		if err != nil {
			RenderError(w, r, err)
			return
		}
	}

	// Check for new rows
	for _, rowNumber := range r.Form["row[]"] {
		row := models.InvoiceRow{
			Description: r.FormValue(fmt.Sprintf("description[%s]", rowNumber)),
		}
		if row.Description == "" {
			fmt.Printf("skipping %s\n", rowNumber)
			continue
		}

		row.Cost, err = decimal.NewFromString(r.FormValue(fmt.Sprintf("cost[%s]", rowNumber)))
		if err != nil {
			RenderError(w, r, err)
			return
		}
		row.IsRotRut, _ = strconv.ParseBool(r.FormValue(fmt.Sprintf("is_rot_rut[%s]", rowNumber)))

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
