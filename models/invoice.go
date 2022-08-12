package models

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"github.com/yzzyx/faktura-pdf/sqlx"
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
	IsOffered      bool
	IsInvoiced     bool
	IsPaid         bool
	IsDeleted      bool
	RutApplicable  bool // Is ROT/RUT applicable for this invoice?
	AdditionalInfo string
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
	ListPaid  bool
	OrderBy   string
	Direction string
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
	combined.Customer = totals.Customer.Add(rowTotals.Customer)
	combined.ROTRUT = totals.ROTRUT.Add(rowTotals.ROTRUT)
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
			vatAmount = priceInclRUT.Sub(priceInclRUT.Div(decimal.NewFromInt(1).Add(row.VAT.Amount())))
		}
	} else {
		totals.Total = totals.Excl
		totals.PPU = totals.PPUExcl
	}

	switch row.VAT {
	case 0:
		totals.VAT25 = vatAmount
	case 1:
		totals.VAT12 = vatAmount
	case 2:
		totals.VAT6 = vatAmount
	}

	return totals
}

func InvoiceGet(ctx context.Context, id int) (Invoice, error) {
	var inv Invoice

	query := `
	SELECT invoice.id,
		invoice.number,
		invoice.name,
		date_created,
		date_paid,
		date_invoiced,
		date_due,
		rut_applicable,
		is_offered,
		is_invoiced,
		is_paid,
		is_deleted,
		additional_info,
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
	INNER JOIN customer ON customer.id = invoice.customer_id
	WHERE invoice.id = $1`

	c := sqlx.NewPgxPool(dbpool)
	err := c.Get(ctx, &inv, query, id)
	if err != nil {
		return inv, err
	}

	err = c.Select(ctx, &inv.Rows, "SELECT id, row_order, description, cost, count, unit, vat, is_rot_rut, rot_rut_service_type, rot_rut_hours, cost*count AS total FROM invoice_row WHERE invoice_id = $1 ORDER BY row_order", inv.ID)
	if err != nil {
		return inv, err
	}

	return inv, nil
}

func InvoiceSave(ctx context.Context, invoice Invoice) (int, error) {
	if invoice.ID > 0 {
		_, err := dbpool.Exec(ctx, `UPDATE invoice SET 
name = $2,
customer_id = $3,
is_invoiced = $4,
is_offered = $5,
is_paid = $6,
additional_info = $7,
date_invoiced = $8,
date_due = $9,
date_paid = $10,
rut_applicable = $11
WHERE id = $1`, invoice.ID,
			invoice.Name,
			invoice.Customer.ID,
			invoice.IsInvoiced,
			invoice.IsOffered,
			invoice.IsPaid,
			invoice.AdditionalInfo,
			invoice.DateInvoiced,
			invoice.DateDue,
			invoice.DatePaid,
			invoice.RutApplicable)
		return invoice.ID, err
	}

	query := `INSERT INTO invoice (number, name, customer_id, rut_applicable) VALUES($1, $2, $3, $4) RETURNING id`
	err := dbpool.QueryRow(ctx, query, invoice.Number, invoice.Name, invoice.Customer.ID, invoice.RutApplicable).Scan(&invoice.ID)
	if err != nil {
		return 0, err
	}
	return invoice.ID, nil
}

func InvoiceList(ctx context.Context, f InvoiceFilter) ([]Invoice, error) {
	var invoices []Invoice
	query := `SELECT
invoice.id,
invoice.number,
invoice.name,
customer.email AS "customer.email",
date_created,
date_paid,
date_invoiced,
date_due,
rut_applicable,
is_offered,
is_invoiced,
is_paid,
is_deleted,
COALESCE((SELECT SUM(r.cost) FROM invoice_row r WHERE r.invoice_id = invoice.id), 0) AS total_sum
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

	filterStrings := []string{}

	if f.ListPaid {
		filterStrings = append(filterStrings, "is_paid")
	} else {
		filterStrings = append(filterStrings, "not is_paid")
	}

	query += " WHERE " + strings.Join(filterStrings, " AND ")
	query += fmt.Sprintf(" ORDER BY %s %s", orderBy, f.Direction)
	c := sqlx.NewPgxPool(dbpool)
	err := c.Select(ctx, &invoices, query)
	if err != nil {
		return nil, err
	}

	for k := range invoices {
		inv := invoices[k]
		err = c.Select(ctx, &inv.Rows, "SELECT id, row_order, description, cost, is_rot_rut, rot_rut_service_type, rot_rut_hours FROM invoice_row WHERE invoice_id = $1 ORDER BY row_order", inv.ID)
		if err != nil {
			return nil, err
		}

	}
	return invoices, nil
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
	_, err := dbpool.Exec(ctx, query,
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
		return err
	}

	return nil
}

func InvoiceRowAdd(ctx context.Context, invoiceID int, row InvoiceRow) error {
	_, err := dbpool.Exec(ctx, `INSERT INTO invoice_row (invoice_id, row_order, description, cost, count, unit, vat, is_rot_rut, rot_rut_service_type)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		invoiceID, row.RowOrder, row.Description, row.Cost, row.Count, row.Unit, row.VAT, row.IsRotRut, row.RotRutServiceType)
	if err != nil {
		return err
	}

	return nil
}

func InvoiceRowRemove(ctx context.Context, invoiceID int, rowID int) error {
	_, err := dbpool.Exec(ctx, `DELETE FROM invoice_row WHERE invoice_id = $1 AND  id = $2`, invoiceID, rowID)
	if err != nil {
		return err
	}

	return nil
}

func InvoiceGetNextNumber(ctx context.Context) (int, error) {
	var num int
	row := dbpool.QueryRow(ctx, `
SELECT COALESCE(MAX(number), 0) + 1 FROM invoice 
`)

	err := row.Scan(&num)
	return num, err
}
