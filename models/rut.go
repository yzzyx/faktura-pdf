package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type RUTStatus int

const (
	RUTStatusPending  RUTStatus = 0
	RUTStatusSent     RUTStatus = 1
	RUTStatusPaid     RUTStatus = 2
	RUTStatusRejected RUTStatus = 3
)

var rutStatusString = map[RUTStatus]string{
	RUTStatusPending:  "skall skickas in",
	RUTStatusSent:     "inskickad",
	RUTStatusPaid:     "betalad",
	RUTStatusRejected: "avslagen",
}

func (r RUTStatus) String() string {
	return rutStatusString[r]
}

type RUTType int

const (
	RUTTypeRUT RUTType = 0
	RUTTypeROT RUTType = 1
)

var rutTypeString = map[RUTType]string{
	RUTTypeRUT: "RUT",
	RUTTypeROT: "ROT",
}

func (r RUTType) String() string {
	return rutTypeString[r]
}

type RUT struct {
	ID           int
	Type         RUTType
	Invoice      Invoice
	Status       RUTStatus
	RequestedSum *int
	DateSent     *time.Time
	DatePaid     *time.Time
}

type RUTFilter struct {
	ID           int
	FilterStatus []RUTStatus
	OrderBy      string
	Direction    string
	InvoiceID    int
	CompanyID    int
	Type         *RUTType

	IncludeInvoice bool
}

func RUTSave(ctx context.Context, rut RUT) (int, error) {
	tx := getContextTx(ctx)
	if rut.ID > 0 {
		_, err := tx.Exec(ctx, `UPDATE rut_requests SET 
status = $2,
date_sent = $3,
date_paid = $4,
requested_sum = $5
WHERE id = $1`, rut.ID,
			rut.Status,
			rut.DateSent,
			rut.DatePaid,
			rut.RequestedSum)

		// If rows are supplied, and rot_rut_hours are set, we update those as well
		for _, row := range rut.Invoice.Rows {
			if row.RotRutHours == nil {
				continue
			}

			err = InvoiceRowUpdate(ctx, row)
			if err != nil {
				return rut.ID, err
			}
		}

		return rut.ID, err
	}

	query := `INSERT INTO rut_requests (invoice_id, status, type) VALUES($1, $2, $3) RETURNING id`
	err := tx.QueryRow(ctx, query, rut.Invoice.ID, rut.Status, rut.Type).Scan(&rut.ID)
	if err != nil {
		return 0, err
	}

	return rut.ID, nil
}

func RUTGet(ctx context.Context, f RUTFilter) (RUT, error) {
	var rutRequest RUT

	lst, err := RUTList(ctx, f)
	if err != nil {
		return rutRequest, err
	}

	if len(lst) == 0 {
		return rutRequest, sql.ErrNoRows
	}

	if len(lst) > 1 {
		return rutRequest, errors.New("too many rows returned")
	}

	rutRequest = lst[0]

	return rutRequest, nil
}

func RUTList(ctx context.Context, f RUTFilter) ([]RUT, error) {
	var rutRequests []RUT
	query := `SELECT
rut_requests.id,
rut_requests.type,
rut_requests.status,
rut_requests.date_sent,
rut_requests.date_paid,
rut_requests.requested_sum,
invoice.id AS "invoice.id",
invoice.number AS "invoice.number",
invoice.name AS "invoice.name",
customer.email AS "invoice.customer.email"
FROM rut_requests
INNER JOIN invoice ON invoice.id = rut_requests.invoice_id
INNER JOIN customer ON customer.id = invoice.customer_id
`

	orderMap := map[string]string{
		"status":         "rot_rut.status",
		"name":           "invoice.name",
		"customer_email": "customer.email",
		"date_sent":      "date_sent",
		"date_paid":      "date_paid",
	}

	orderBy, ok := orderMap[f.OrderBy]
	if !ok {
		orderBy = "invoice.number"
	}

	if strings.ToUpper(f.Direction) != "DESC" {
		f.Direction = "ASC"
	}

	filterStrings := []string{}

	if len(f.FilterStatus) > 0 {
		var fs []string
		for _, v := range f.FilterStatus {
			fs = append(fs, strconv.Itoa(int(v)))
		}
		filterStrings = append(filterStrings, fmt.Sprintf(`rut_requests.status IN (%s)`, strings.Join(fs, ",")))
	}

	if f.InvoiceID > 0 {
		filterStrings = append(filterStrings, "rut_requests.invoice_id = :invoice_id")
	}

	if f.Type != nil {
		filterStrings = append(filterStrings, "rut_requests.type = :type")
	}

	if f.ID > 0 {
		filterStrings = append(filterStrings, "rut_requests.id = :id")
	}

	if f.CompanyID > 0 {
		filterStrings = append(filterStrings, "invoice.company_id = :company_id")
	}

	if len(filterStrings) > 0 {
		query += " WHERE " + strings.Join(filterStrings, " AND ")
	}

	query += fmt.Sprintf(" ORDER BY %s %s", orderBy, f.Direction)

	tx := getContextTx(ctx)
	rows, err := tx.NamedQuery(ctx, query, f)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var r RUT
		err = rows.StructScan(&r)
		if err != nil {
			return nil, err
		}

		rutRequests = append(rutRequests, r)
	}

	if f.IncludeInvoice {
		for k := range rutRequests {
			rutRequests[k].Invoice, err = InvoiceGet(ctx, InvoiceFilter{ID: rutRequests[k].Invoice.ID})
			if err != nil {
				return nil, err
			}
		}
	}

	return rutRequests, nil
}
