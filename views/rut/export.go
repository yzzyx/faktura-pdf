package rut

import (
	"encoding/xml"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"github.com/yzzyx/faktura-pdf/models"
	"github.com/yzzyx/faktura-pdf/rotrut"
	"github.com/yzzyx/faktura-pdf/views"
)

// Export is the view-handler for creating a SKV ROT/RUT request
type Export struct {
	views.View
}

// NewExport creates a new handler for creating a SKV ROT/RUT request
func NewExport() *Export {
	return &Export{}
}

// HandleGet creates an xml export file
func (v *Export) HandleGet() error {
	id := v.URLParamInt("id")

	rutRequest, err := models.RUTGet(v.Ctx, id)
	if err != nil {
		return err
	}

	var rutBegaran *rotrut.HushallBegaranTYPE

	if rutRequest.Type == models.RUTTypeRUT {
		rutBegaran = &rotrut.HushallBegaranTYPE{
			Arenden: []rotrut.HushallArendeTYPE{
				{
					Kopare:          rotrut.PeOrgNrTYPE(rutRequest.Invoice.Customer.PNR),
					BegartBelopp:    rotrut.BeloppTYPE(*rutRequest.RequestedSum),
					FakturaNr:       rotrut.FakturaNrTYPE(strconv.Itoa(rutRequest.Invoice.Number)),
					BetalningsDatum: rotrut.DatumTYPE(rutRequest.Invoice.DatePaid.Format("2006-01-02")),
					UtfortArbete: &rotrut.ArendeUtfortArbeteRutTYPE{
						Stadning:                 &rotrut.TimmarMaterialTYPE{},
						KladOchTextilvard:        &rotrut.TimmarMaterialTYPE{},
						Snoskottning:             &rotrut.TimmarMaterialTYPE{},
						Tradgardsarbete:          &rotrut.TimmarMaterialTYPE{},
						Barnpassning:             &rotrut.TimmarMaterialTYPE{},
						Personligomsorg:          &rotrut.TimmarMaterialTYPE{},
						Flyttjanster:             &rotrut.TimmarMaterialTYPE{},
						ItTjanster:               &rotrut.TimmarMaterialTYPE{},
						ReparationAvVitvaror:     &rotrut.TimmarMaterialTYPE{},
						Moblering:                &rotrut.TimmarMaterialTYPE{},
						TillsynAvBostad:          &rotrut.TimmarMaterialTYPE{},
						TransportTillForsaljning: nil, // Not implemented
						TvattVidTvattinrattning:  nil, // Not implemented
					},
				},
			},
		}
		rutBegaran.SetXMLNS()

		arende := &rutBegaran.Arenden[0]
		ua := arende.UtfortArbete
		arbetsKostnad := decimal.Decimal{}
		ovrigKostnad := decimal.Decimal{}

		for _, r := range rutRequest.Invoice.Rows {
			if !r.IsRotRut || r.RotRutServiceType == nil {
				ovrigKostnad = ovrigKostnad.Add(r.Total)
				continue
			}

			var hours rotrut.AntalTimmarTYPE
			if r.Unit == 2 {
				hours = rotrut.AntalTimmarTYPE(r.Count.IntPart())
			} else if r.RotRutHours != nil {
				hours = rotrut.AntalTimmarTYPE(*r.RotRutHours)
			}

			if hours == 0 {
				continue
			}

			arbetsKostnad = arbetsKostnad.Add(r.Total)
			switch *r.RotRutServiceType {
			case models.RUTServiceTypeStadning:
				ua.Stadning.AntalTimmar += hours
			case models.RUTServiceTypeKladOchTextilvard:
				ua.KladOchTextilvard.AntalTimmar += hours
			case models.RUTServiceTypeSnoskottning:
				ua.Snoskottning.AntalTimmar += hours
			case models.RUTServiceTypeTradgardsarbete:
				ua.Tradgardsarbete.AntalTimmar += hours
			case models.RUTServiceTypeBarnpassning:
				ua.Barnpassning.AntalTimmar += hours
			case models.RUTServiceTypePersonligomsorg:
				ua.Personligomsorg.AntalTimmar += hours
			case models.RUTServiceTypeFlyttjanster:
				ua.Flyttjanster.AntalTimmar += hours
			case models.RUTServiceTypeITTjanster:
				ua.ItTjanster.AntalTimmar += hours
			case models.RUTServiceTypeReparationAvVitvaror:
				ua.ReparationAvVitvaror.AntalTimmar += hours
			case models.RUTServiceTypeMoblering:
				ua.Moblering.AntalTimmar += hours
			case models.RUTServiceTypeTillsynAvBostad:
				ua.TillsynAvBostad.AntalTimmar += hours
			case models.RUTServiceTypeTransportTillForsaljning:
				return errors.New("export av transport till försäljning ej implementerat")
			case models.RUTServiceTypeTvattVidTvattinrattning:
				return errors.New("export av tvätt vid tvättinrättning ej implementerat")
			}
		}

		arende.PrisForArbete = arbetsKostnad.StringFixedBank(0)
		arende.BetaltBelopp = rotrut.BeloppTYPE(arbetsKostnad.Sub(decimal.NewFromInt(int64(*rutRequest.RequestedSum))).IntPart())
		arende.Ovrigkostnad = rotrut.OvrigKostnadTYPE(ovrigKostnad.IntPart())
	} else {
		return errors.New("export av ROT-ärenden ej implementerad")
	}

	now := time.Now()
	begaran := rotrut.NewBegaran(fmt.Sprintf("%s-%d", rutRequest.Type, rutRequest.Invoice.Number))
	begaran.HushallBegaran = rutBegaran

	output, err := xml.MarshalIndent(begaran, "", "  ")
	if err != nil {
		return err
	}

	name := fmt.Sprintf("RUT-%d-%s.xml", rutRequest.Invoice.Number, now.Format("2006-01-02"))
	name = strings.ReplaceAll(name, " ", "_")

	headers := v.ResponseHeaders()
	headers.Set("Content-Type", "application/xml")
	headers.Set("Content-Disposition", "attachment; filename="+name)
	err = v.RenderBytes([]byte(xml.Header))
	if err != nil {
		return err
	}

	return v.RenderBytes(output)
}
