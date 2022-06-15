package rut

import (
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
	var rut models.RUT

	id := v.URLParamInt("id")

	rut, err = models.RUTGet(v.Ctx, id)
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
	case "paid":
		rut.Status = models.RUTStatusPaid
		rut.DatePaid = &date
	case "rejected":
		rut.Status = models.RUTStatusRejected
	default:
		return errors.New("invalid flag")
	}

	_, err = models.RUTSave(v.Ctx, rut)
	if err != nil {
		return err
	}

	return v.RedirectRoute("rut-view", "id", strconv.Itoa(id))
}

// HandlePost updates the flags of an invoice
func (v *Flag) HandlePost() error {
	return v.HandleGet()
}
