package rut

import (
	"errors"
	"strconv"
	"time"

	"github.com/yzzyx/faktura-pdf/models"
	"github.com/yzzyx/faktura-pdf/views"
)

// Flag is the view-handler for updating RUT flags
type Flag struct {
	views.View
}

// NewFlag creates a new handler for updating RUT flags
func NewFlag() *Flag {
	return &Flag{}
}

// HandlePost updates the flags of an RUT request
func (v *Flag) HandlePost() error {
	var err error
	f := models.RUTFilter{
		ID:             v.URLParamInt("id"),
		CompanyID:      v.Session.Company.ID,
		IncludeInvoice: true,
	}

	if f.ID <= 0 {
		return views.ErrBadRequest
	}

	rutRequest, err := models.RUTGet(v.Ctx, f)
	if err != nil {
		return err
	}

	flag := v.FormValueString("flag")
	date := time.Now()
	if v.FormValueExists("date") {
		if v, err := time.Parse("2006-01-02", v.FormValueString("date")); err == nil {
			date = v
		}
	}

	switch flag {
	case "sent":
		rutRequest.Status = models.RUTStatusSent
		rutRequest.DateSent = &date
	case "paid":
		rutRequest.Status = models.RUTStatusPaid
		rutRequest.DatePaid = &date
		receivedAmount := v.FormValueInt("amount")
		rutRequest.ReceivedSum = &receivedAmount
	case "rejected":
		rutRequest.Status = models.RUTStatusRejected
		rutRequest.DatePaid = &date
	default:
		return errors.New("invalid flag")
	}

	_, err = models.RUTSave(v.Ctx, rutRequest)
	if err != nil {
		return err
	}

	return v.RedirectRoute("rut-view", "id", strconv.Itoa(f.ID))
}
