package models

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"github.com/yzzyx/zerr"
)

// InvoiceStatus is used for mutually exclusive flags
type InvoiceStatus int

const (
	InvoiceStatusInitial InvoiceStatus = iota
	InvoiceStatusOffered
	InvoiceStatusAccepted
	InvoiceStatusRejected
)

type Invoice struct {
	ID             int
	Number         int
	Name           string
	DateCreated    time.Time
	DatePaid       *time.Time
	DateInvoiced   *time.Time
	DateDue        *time.Time
	TotalSum       decimal.Decimal
	Customer       Customer
	Rows           []InvoiceRow
	IsInvoiced     bool
	IsPaid         bool
	IsDeleted      bool
	RutApplicable  bool // Is ROT/RUT applicable for this invoice?
	AdditionalInfo string
	Status         InvoiceStatus

	IsOffer bool // Is this an offer, instead of an invoice?
	OfferID *int // Was this invoice created from an offer?

	Company Company
}

type UnitType int

var unitTypeString = map[int]string{
	0: "-",
	1: "st",
	2: "timmar",
	3: "dagar",
}

func (u UnitType) Validate() bool {
	_, ok := unitTypeString[int(u)]
	return ok
}

func (u UnitType) String() string {
	v := unitTypeString[int(u)]
	return v
}

type VATType int

var vatTypeString = map[int]string{
	0: "25 %",
	1: "12 %",
	2: "6 %",
	3: "0 %",
}

var vatTypeAmount = map[VATType]decimal.Decimal{
	0: decimal.NewFromFloat(0.25),
	1: decimal.NewFromFloat(0.12),
	2: decimal.NewFromFloat(0.6),
	3: decimal.NewFromFloat(0),
}

func (u VATType) Validate() bool {
	_, ok := vatTypeString[int(u)]
	return ok
}

func (u VATType) String() string {
	v := vatTypeString[int(u)]
	return v
}

func (u VATType) Amount() decimal.Decimal {
	v := vatTypeAmount[u]
	return v
}

type ROTRUTServiceType int

const (
	ROTServiceTypeBygg ROTRUTServiceType = iota
	ROTServiceTypeEl
	ROTServiceTypeGlasPlatarbete
	ROTServiceTypeMarkDraneringsarbete
	ROTServiceTypeMurning
	ROTServiceTypeMalningTapetsering
	ROTServiceTypeVVS

	RUTServiceTypeStadning
	RUTServiceTypeKladOchTextilvard
	RUTServiceTypeSnoskottning
	RUTServiceTypeTradgardsarbete
	RUTServiceTypeBarnpassning
	RUTServiceTypePersonligomsorg
	RUTServiceTypeFlyttjanster
	RUTServiceTypeITTjanster
	RUTServiceTypeReparationAvVitvaror
	RUTServiceTypeMoblering
	RUTServiceTypeTillsynAvBostad
	RUTServiceTypeTransportTillForsaljning
	RUTServiceTypeTvattVidTvattinrattning
)

var ROTServices = map[ROTRUTServiceType]string{
	ROTServiceTypeBygg:                 "Bygg",
	ROTServiceTypeEl:                   "El",
	ROTServiceTypeGlasPlatarbete:       "Glas och plåt",
	ROTServiceTypeMarkDraneringsarbete: "Mark- och dräneringsarbete",
	ROTServiceTypeMurning:              "Murning",
	ROTServiceTypeMalningTapetsering:   "Tapetsering",
	ROTServiceTypeVVS:                  "VVS",
}

var RUTServices = map[ROTRUTServiceType]string{
	RUTServiceTypeStadning:                 "Städning",
	RUTServiceTypeKladOchTextilvard:        "Kläd- och textilvård",
	RUTServiceTypeSnoskottning:             "Snöskottning",
	RUTServiceTypeTradgardsarbete:          "Trädgårdsarbete",
	RUTServiceTypeBarnpassning:             "Barnpassning",
	RUTServiceTypePersonligomsorg:          "Personling omsorg",
	RUTServiceTypeFlyttjanster:             "Flyttjänster",
	RUTServiceTypeITTjanster:               "IT-tjänster",
	RUTServiceTypeReparationAvVitvaror:     "Reparation av vitvaror",
	RUTServiceTypeMoblering:                "Möblering",
	RUTServiceTypeTillsynAvBostad:          "Tillsyn av bostad",
	RUTServiceTypeTransportTillForsaljning: "Transport till försäljning",
	RUTServiceTypeTvattVidTvattinrattning:  "Tvätt vid tvättinrättning",
}

func (s ROTRUTServiceType) String() string {
	if v, ok := ROTServices[s]; ok {
		return v
	}
	return RUTServices[s]
}

func (s ROTRUTServiceType) IsROT() bool {
	_, ok := ROTServices[s]
	return ok
}

func (s ROTRUTServiceType) IsRUT() bool {
	_, ok := RUTServices[s]
	return ok
}

type InvoiceRow struct {
	ID          int
	RowOrder    int             `json:"row_order"`
	Description string          `json:"description"`
	Cost        decimal.Decimal `json:"cost"`
	Count       decimal.Decimal `json:"count"`
	Unit        UnitType        `json:"unit"`
	VAT         VATType         `json:"vat"`

	// Fields used for ROT/RUT
	IsRotRut          bool               `json:"is_rot_rut"`
	RotRutServiceType *ROTRUTServiceType `json:"rot_rut_service_type"`
	RotRutHours       *int               `json:"rot_rut_hours"` // used when a row has a fixed price

	// Calculated fields
	Total decimal.Decimal
}

type InvoiceFilter struct {
	ID        int
	CompanyID int

	ListOffers bool // false - list invoices, true, list offers
	FilterPaid int  // 0 - no filter, 1 - only paid, 2 - only unpaid
	OrderBy    string
	Direction  string
	Status     []InvoiceStatus // Accepted statuses

	IncludeCompany bool
}

type InvoiceTotals struct {
	Total    decimal.Decimal // Including or Excluding VAT and ROT/RUT, depending on flag
	Incl     decimal.Decimal // Including VAT
	Excl     decimal.Decimal // Excluding VAT
	VAT25    decimal.Decimal // 25% VAT
	VAT12    decimal.Decimal // 12% VAT
	VAT6     decimal.Decimal // 6% VAT
	Customer decimal.Decimal // Amount customer pays
	ROTRUT   decimal.Decimal // Amount of ROT/RUT

	TotalVAT25 decimal.Decimal // Total excl for 25% VAT
	TotalVAT12 decimal.Decimal // Total excl for 12% VAT
	TotalVAT6  decimal.Decimal // Total excl for 6% VAT

	// Total sums for ROT/RUT only
	ROTRUTTotals struct {
		Incl  decimal.Decimal // ROT/RUT incl VAT
		Excl  decimal.Decimal // ROT/RUT excl VAT
		VAT25 decimal.Decimal // ROT/RUT 25% VAT
		VAT12 decimal.Decimal // ROT/RUT 12% VAT
		VAT6  decimal.Decimal // ROT/RUT 6% VAT
	}

	// For single rows
	PPU           decimal.Decimal // Price-Per-Unit including or excluding VAT and ROT/RUT, depending on flag
	PPUIncl       decimal.Decimal // Price-Per-Unit Including VAT
	PPUExcl       decimal.Decimal // Price-Per-Unit excluding VAT
	ROTRUTPerUnit decimal.Decimal // ROT/RUT per unit
}

func (totals InvoiceTotals) Add(rowTotals InvoiceTotals) (combined InvoiceTotals) {
	combined.Total = totals.Total.Add(rowTotals.Total)
	combined.Incl = totals.Incl.Add(rowTotals.Incl)
	combined.Excl = totals.Excl.Add(rowTotals.Excl)
	combined.VAT25 = totals.VAT25.Add(rowTotals.VAT25)
	combined.VAT12 = totals.VAT12.Add(rowTotals.VAT12)
	combined.VAT6 = totals.VAT6.Add(rowTotals.VAT6)
	combined.TotalVAT25 = totals.TotalVAT25.Add(rowTotals.TotalVAT25)
	combined.TotalVAT12 = totals.TotalVAT12.Add(rowTotals.TotalVAT12)
	combined.TotalVAT6 = totals.TotalVAT6.Add(rowTotals.TotalVAT6)
	combined.Customer = totals.Customer.Add(rowTotals.Customer)
	combined.ROTRUT = totals.ROTRUT.Add(rowTotals.ROTRUT)

	combined.ROTRUTTotals.Incl = totals.ROTRUTTotals.Incl.Add(rowTotals.ROTRUTTotals.Incl)
	combined.ROTRUTTotals.Excl = totals.ROTRUTTotals.Excl.Add(rowTotals.ROTRUTTotals.Excl)
	combined.ROTRUTTotals.VAT25 = totals.ROTRUTTotals.VAT25.Add(rowTotals.ROTRUTTotals.VAT25)
	combined.ROTRUTTotals.VAT12 = totals.ROTRUTTotals.VAT12.Add(rowTotals.ROTRUTTotals.VAT12)
	combined.ROTRUTTotals.VAT6 = totals.ROTRUTTotals.VAT6.Add(rowTotals.ROTRUTTotals.VAT6)
	return combined
}

func (i *Invoice) Totals(IncludeVAT, IncludeROTRUT bool) (totals InvoiceTotals) {
	for _, row := range i.Rows {
		rowTotals := row.Totals(IncludeVAT, IncludeROTRUT)
		totals = totals.Add(rowTotals)
	}

	return totals
}

func (row *InvoiceRow) Totals(IncludeVAT bool, IncludeROTRUT bool) (totals InvoiceTotals) {
	priceInclRUT := row.Cost
	totals.PPUIncl = row.Cost

	if row.IsRotRut && row.RotRutServiceType != nil {
		if row.RotRutServiceType.IsROT() {
			priceInclRUT = row.Cost.Mul(decimal.NewFromFloat(0.7))
			totals.ROTRUT = totals.ROTRUT.Add(row.Cost.Mul(decimal.NewFromFloat(0.3)).Mul(row.Count))
		} else {
			priceInclRUT = row.Cost.Mul(decimal.NewFromFloat(0.5))
			totals.ROTRUT = totals.ROTRUT.Add(priceInclRUT.Mul(row.Count))
		}

		totals.ROTRUTTotals.Incl = totals.ROTRUT
		totals.ROTRUTTotals.Excl = totals.ROTRUT.Div(decimal.NewFromInt(1).Add(row.VAT.Amount()))
		totals.ROTRUTPerUnit = row.Cost.Mul(row.Count).Sub(priceInclRUT)
	}

	totals.Customer = totals.Customer.Add(priceInclRUT.Mul(row.Count))
	totals.Incl = totals.Incl.Add(row.Total)

	priceExcl := row.Cost.Div(decimal.NewFromInt(1).Add(row.VAT.Amount()))
	totals.Excl = priceExcl.Mul(row.Count)
	totals.PPUExcl = priceExcl
	vatAmount := totals.Incl.Sub(totals.Excl)

	if IncludeVAT {
		totals.Total = totals.Incl
		totals.PPU = totals.PPUIncl

		if IncludeROTRUT {
			totals.Total = totals.Customer
			totals.PPU = priceInclRUT
			vatAmount = priceInclRUT.Sub(priceInclRUT.Div(decimal.NewFromInt(1).Add(row.VAT.Amount()))).Mul(row.Count)
		}
	} else {
		totals.Total = totals.Excl
		totals.PPU = totals.PPUExcl
	}

	switch row.VAT {
	case 0:
		totals.VAT25 = vatAmount
		totals.TotalVAT25 = totals.Excl
		totals.ROTRUTTotals.VAT25 = totals.ROTRUTTotals.Incl.Sub(totals.ROTRUTTotals.Excl)
	case 1:
		totals.VAT12 = vatAmount
		totals.TotalVAT12 = totals.Excl
		totals.ROTRUTTotals.VAT12 = totals.ROTRUTTotals.Incl.Sub(totals.ROTRUTTotals.Excl)
	case 2:
		totals.VAT6 = vatAmount
		totals.TotalVAT6 = totals.Excl
		totals.ROTRUTTotals.VAT6 = totals.ROTRUTTotals.Incl.Sub(totals.ROTRUTTotals.Excl)
	}

	return totals
}

func InvoiceGet(ctx context.Context, filter InvoiceFilter) (Invoice, error) {
	var inv Invoice

	lst, err := InvoiceList(ctx, filter)
	if err != nil {
		return inv, err
	}

	if len(lst) == 0 {
		return inv, sql.ErrNoRows
	}

	if len(lst) > 1 {
		return inv, zerr.Wrap(errTooManyRows).WithAny("filter", filter)
	}

	inv = lst[0]

	return inv, nil
}

func InvoiceAddAttachment(ctx context.Context, inv Invoice, f File) error {
	var err error
	if f.ID == 0 {
		f.ID, err = FileAdd(ctx, f)
		if err != nil {
			return err
		}
	}

	tx := getContextTx(ctx)

	query := `INSERT INTO invoice_attachments (invoice_id, file_id) VALUES ($1, $2)`

	_, err = tx.Exec(ctx, query, inv.ID, f.ID)
	if err != nil {
		return zerr.Wrap(err).WithString("query", query).WithInt("invoice_id", inv.ID).WithInt("file_id", f.ID)
	}
	return nil
}

func InvoiceSave(ctx context.Context, invoice Invoice) (int, error) {
	tx := getContextTx(ctx)

	if invoice.ID > 0 {
		query := `UPDATE invoice SET 
name = $2,
customer_id = $3,
is_invoiced = $4,
is_paid = $5,
additional_info = $6,
date_invoiced = $7,
date_due = $8,
date_paid = $9,
rut_applicable = $10,
is_deleted = $11,
status = $12
WHERE id = $1`
		_, err := tx.Exec(ctx, query, invoice.ID,
			invoice.Name,
			invoice.Customer.ID,
			invoice.IsInvoiced,
			invoice.IsPaid,
			invoice.AdditionalInfo,
			invoice.DateInvoiced,
			invoice.DateDue,
			invoice.DatePaid,
			invoice.RutApplicable,
			invoice.IsDeleted,
			invoice.Status)
		if err != nil {
			return 0, zerr.Wrap(err).WithString("query", query).WithAny("invoice", invoice)
		}
		return invoice.ID, nil
	}

	query := `INSERT INTO invoice (number, name, customer_id, rut_applicable, company_id, is_offer, offer_id, status) VALUES($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`
	err := tx.QueryRow(ctx, query, invoice.Number, invoice.Name, invoice.Customer.ID, invoice.RutApplicable, invoice.Company.ID, invoice.IsOffer, invoice.OfferID, invoice.Status).Scan(&invoice.ID)
	if err != nil {
		return 0, zerr.Wrap(err).WithString("query", query).WithAny("invoice", invoice)
	}
	return invoice.ID, nil
}

func invoiceBuildQuery(q string, f InvoiceFilter) string {
	filterStrings := []string{"not is_deleted"}

	if f.FilterPaid == 1 {
		filterStrings = append(filterStrings, "is_paid")
	} else if f.FilterPaid == 2 {
		filterStrings = append(filterStrings, "not is_paid")
	}

	if f.ListOffers {
		filterStrings = append(filterStrings, "is_offer")
	} else {
		filterStrings = append(filterStrings, "not is_offer")
	}

	if f.ID > 0 {
		filterStrings = append(filterStrings, "invoice.id = :id")
	}

	if f.CompanyID > 0 {
		filterStrings = append(filterStrings, "invoice.company_id = :company_id")
	}

	if len(f.Status) > 0 {
		statusFilter := make([]string, len(f.Status))
		for k, v := range f.Status {
			statusFilter[k] = strconv.Itoa(int(v))
		}
		filterStrings = append(filterStrings, "invoice.status IN ("+strings.Join(statusFilter, ",")+")")
	}

	q += " WHERE " + strings.Join(filterStrings, " AND ")
	return q
}

func InvoiceList(ctx context.Context, f InvoiceFilter) ([]Invoice, error) {
	var invoices []Invoice
	query := `SELECT
		invoice.id,
		invoice.number,
		invoice.name,
		date_created,
		date_paid,
		date_invoiced,
		date_due,
		rut_applicable,
		is_invoiced,
		is_paid,
		is_deleted,
		is_offer,
		status,
		offer_id,
		additional_info,
		invoice.company_id AS "company.id",
		customer.id AS "customer.id",
		customer.name AS "customer.name",
		customer.email AS "customer.email",
		customer.address1 AS "customer.address1",
		customer.address2 AS "customer.address2",
		customer.postcode AS "customer.postcode",
		customer.city AS "customer.city",
		customer.pnr AS "customer.pnr",
		customer.telephone AS "customer.telephone",
		COALESCE((SELECT SUM(r.cost*r.count) FROM invoice_row r WHERE r.invoice_id = invoice.id), 0) AS total_sum
FROM invoice
INNER JOIN customer ON customer.id = invoice.customer_id`

	orderMap := map[string]string{
		"number":         "invoice.number",
		"name":           "invoice.name",
		"customer_email": "customer.email",
		"date_created":   "date_created",
		"date_paid":      "date_paid",
		"date_due":       "date_due",
		"total_sum":      "total_sum",
	}

	orderBy, ok := orderMap[f.OrderBy]
	if !ok {
		orderBy = "invoice.number"
	}

	if strings.ToUpper(f.Direction) != "DESC" {
		f.Direction = "ASC"
	}

	query = invoiceBuildQuery(query, f)
	query += fmt.Sprintf(" ORDER BY %s %s", orderBy, f.Direction)

	tx := getContextTx(ctx)
	rows, err := tx.NamedQuery(ctx, query, f)
	if err != nil {
		return nil, zerr.Wrap(err).WithString("query", query).WithAny("filter", f)
	}

	for rows.Next() {
		var invoice Invoice
		err = rows.StructScan(&invoice)
		if err != nil {
			return nil, zerr.Wrap(err).WithString("query", query).WithAny("filter", f)
		}

		invoices = append(invoices, invoice)
	}

	for k := range invoices {
		inv := &invoices[k]
		query := "SELECT id, row_order, description, cost, count, unit, vat, is_rot_rut, rot_rut_service_type, rot_rut_hours, cost*count AS total FROM invoice_row WHERE invoice_id = $1 ORDER BY row_order"
		err = tx.Select(ctx, &inv.Rows, query, inv.ID)
		if err != nil {
			return nil, zerr.Wrap(err).WithString("query", query).WithInt("invoice.ID", inv.ID)
		}

		if f.IncludeCompany {
			inv.Company, err = CompanyGet(ctx, CompanyFilter{ID: inv.Company.ID})
			if err != nil {
				return nil, err
			}
		}
	}
	return invoices, nil
}

// InvoiceCount returns the number of entries matching the filter
func InvoiceCount(ctx context.Context, f InvoiceFilter) (int, error) {
	var count int

	query := `SELECT
		COUNT(invoice.id)
FROM invoice
INNER JOIN customer ON customer.id = invoice.customer_id`
	query = invoiceBuildQuery(query, f)

	tx := getContextTx(ctx)
	rows, err := tx.NamedQuery(ctx, query, f)
	if err != nil {
		return 0, zerr.Wrap(err).WithString("query", query)
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&count)
		if err != nil {
			return 0, zerr.Wrap(err).WithString("query", query)
		}
	}
	return count, nil
}

func InvoiceRowUpdate(ctx context.Context, row InvoiceRow) error {
	query := `
	UPDATE invoice_row SET row_order = $2,
		description = $3,
		cost = $4,
		count = $5,
		unit = $6,
		vat = $7,
		is_rot_rut = $8,
		rot_rut_service_type = $9,
		rot_rut_hours = $10
	WHERE id = $1
`
	tx := getContextTx(ctx)
	_, err := tx.Exec(ctx, query,
		row.ID,
		row.RowOrder,
		row.Description,
		row.Cost,
		row.Count,
		row.Unit,
		row.VAT,
		row.IsRotRut,
		row.RotRutServiceType,
		row.RotRutHours)
	if err != nil {
		return zerr.Wrap(err).WithString("query", query).WithAny("row", row)
	}

	return nil
}

func InvoiceRowAdd(ctx context.Context, invoiceID int, row InvoiceRow) error {
	tx := getContextTx(ctx)
	query := `INSERT INTO invoice_row (invoice_id, row_order, description, cost, count, unit, vat, is_rot_rut, rot_rut_service_type)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := tx.Exec(ctx, query,
		invoiceID, row.RowOrder, row.Description, row.Cost, row.Count, row.Unit, row.VAT, row.IsRotRut, row.RotRutServiceType)
	if err != nil {
		return zerr.Wrap(err).WithString("query", query).WithAny("row", row).WithAny("invoice-id", invoiceID)
	}

	return nil
}

func InvoiceRowRemove(ctx context.Context, invoiceID int, rowID int) error {
	tx := getContextTx(ctx)
	query := `DELETE FROM invoice_row WHERE invoice_id = $1 AND  id = $2`
	_, err := tx.Exec(ctx, query, invoiceID, rowID)
	if err != nil {
		return zerr.Wrap(err).WithString("query", query).WithAny("row-id", rowID).WithAny("invoice-id", invoiceID)
	}

	return nil
}
