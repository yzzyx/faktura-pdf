package models

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/shopspring/decimal"
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
	IsRutSent      bool
	IsRutPaid      bool
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

func (u VATType) Validate() bool {
	_, ok := vatTypeString[int(u)]
	return ok
}

func (u VATType) String() string {
	v := vatTypeString[int(u)]
	return v
}

type InvoiceRow struct {
	ID          int
	RowOrder    int             `json:"row_order"`
	Description string          `json:"description"`
	Cost        decimal.Decimal `json:"cost"`
	Count       decimal.Decimal `json:"count"`
	Unit        UnitType        `json:"unit"`
	VAT         VATType         `json:"vat"`
	IsRotRut    bool            `json:"is_rot_rut"`

	Total decimal.Decimal
}

type InvoiceFilter struct {
	IncludePaid bool
	OrderBy     string
	Direction   string
}

func InvoiceGet(ctx context.Context, id int) (Invoice, error) {
	var inv Invoice
	err := pgxscan.Get(ctx, dbpool, &inv, `
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
is_rut_sent,
is_rut_paid,
additional_info,
customer.id AS "customer.id",
customer.name AS "customer.name",
customer.email AS "customer.email",
customer.address1 AS "customer.address1",
customer.address2 AS "customer.address2",
customer.postcode AS "customer.postcode",
customer.city AS "customer.city",
customer.pnr AS "customer.pnr",
COALESCE((SELECT SUM(r.cost*r.count) FROM invoice_row r WHERE r.invoice_id = invoice.id), 0) AS total_sum
FROM invoice
INNER JOIN customer ON customer.id = invoice.customer_id
WHERE invoice.id = $1`, id)
	if err != nil {
		return inv, err
	}

	err = pgxscan.Select(ctx, dbpool, &inv.Rows, "SELECT id, row_order, description, cost, count, unit, vat, is_rot_rut, cost*count AS total FROM invoice_row WHERE invoice_id = $1 ORDER BY row_order", inv.ID)
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
rut_applicable = $11,
is_rut_sent = $12,
is_rut_paid = $13
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
			invoice.RutApplicable,
			invoice.IsRutSent,
			invoice.IsRutPaid)
		return invoice.ID, err
	}

	query := `INSERT INTO invoice (number, name, customer_id, rut_applicable) VALUES($1, $2, $3, $4) RETURNING id`
	err := dbpool.QueryRow(ctx, query, invoice.Number, invoice.Name, invoice.Customer.ID, invoice.RutApplicable).Scan(&invoice.ID)
	if err != nil {
		return 0, err
	}
	return invoice.ID, nil
}

func InvoiceListActive(ctx context.Context, f InvoiceFilter) ([]Invoice, error) {
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
is_rut_sent,
is_rut_paid,
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

	query += fmt.Sprintf(" ORDER BY %s %s", orderBy, f.Direction)
	err := pgxscan.Select(ctx, dbpool, &invoices, query)
	if err != nil {
		return nil, err
	}

	for k := range invoices {
		inv := invoices[k]
		err = pgxscan.Select(ctx, dbpool, &inv.Rows, "SELECT id, row_order, description, cost, is_rot_rut FROM invoice_row WHERE invoice_id = $1 ORDER BY row_order", inv.ID)
		if err != nil {
			return nil, err
		}

	}
	return invoices, nil
}

func InvoiceRowAdd(ctx context.Context, invoiceID int, row InvoiceRow) error {
	_, err := dbpool.Exec(ctx, `INSERT INTO invoice_row (invoice_id, row_order, description, cost, count, unit, vat, is_rot_rut)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		invoiceID, row.RowOrder, row.Description, row.Cost, row.Count, row.Unit, row.VAT, row.IsRotRut)
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
