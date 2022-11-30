package invoice

import (
	"io"
	"mime"
	"path/filepath"
	"strconv"

	"github.com/yzzyx/faktura-pdf/models"
	"github.com/yzzyx/faktura-pdf/views"
	"github.com/yzzyx/zerr"
)

// Attachment is the view-handler for viewing/adding invoice attachments
type Attachment struct {
	IsOffer bool
	views.View
}

// NewAttachment creates a new handler for handling attachments
func NewAttachment(isOffer bool) *Attachment {
	return &Attachment{IsOffer: isOffer}
}

// HandleGet returns the contents of an attachment
func (v *Attachment) HandleGet() error {
	var err error
	invoiceID := v.URLParamInt("id")
	attachmentID := v.URLParamInt("attachment")

	if invoiceID <= 0 || attachmentID <= 0 {
		return views.ErrBadRequest
	}

	lst, err := models.FileList(v.Ctx, models.FileFilter{
		ID:             attachmentID,
		CompanyID:      v.Session.Company.ID,
		InvoiceID:      invoiceID,
		IncludeContent: true,
	})
	if err != nil {
		return err
	}

	if len(lst) == 0 {
		return views.ErrNotFound
	}

	if len(lst) > 1 {
		return views.ErrBadRequest
	}

	f := lst[0]

	headers := v.ResponseHeaders()
	if f.MIMEType != "" {
		headers.Set("Content-Type", f.MIMEType)
	}

	headers.Set("Content-Length", strconv.Itoa(len(f.Contents)))

	return v.RenderBytes(f.Contents)
}

// HandlePost adds an attachment to an invoice
func (v *Attachment) HandlePost() error {
	var err error
	var invoice models.Invoice

	invoiceID := v.URLParamInt("id")
	if invoiceID <= 0 {
		return views.ErrBadRequest
	}

	invoice, err = models.InvoiceGet(v.Ctx, models.InvoiceFilter{ID: invoiceID, CompanyID: v.Session.Company.ID, ListOffers: v.IsOffer})
	if err != nil {
		return err
	}

	files := v.FormFiles("file")

	for _, fileInfo := range files {

		f, err := fileInfo.Open()
		if err != nil {
			return zerr.Wrap(err).WithString("filename", fileInfo.Filename)
		}

		mimeType := mime.TypeByExtension(filepath.Ext(fileInfo.Filename))
		file := models.File{
			Name:      fileInfo.Filename,
			CompanyID: v.Session.Company.ID,
			MIMEType:  mimeType,
			Backend:   nil,
		}

		file.Contents, err = io.ReadAll(f)
		if err != nil {
			return zerr.Wrap(err).WithString("filename", fileInfo.Filename)
		}

		f.Close()

		err = models.InvoiceAddAttachment(v.Ctx, invoice, file)
		if err != nil {
			return err
		}
	}

	return nil
}
