package api

import (
	"github.com/paysuper/paysuper-management-api/database/model"
	"github.com/ttacon/libphonenumber"
	"gopkg.in/go-playground/validator.v9"
)

func ProjectStructValidator(sl validator.StructLevel) {
	p := sl.Current().Interface().(model.ProjectScalar)

	if p.SendNotifyEmail == true && len(p.NotifyEmails) <= 0 {
		sl.ReportError(p.NotifyEmails, "NotifyEmails", "notify_emails", "notify_emails", "")
	}

	if p.OnlyFixedAmounts != true && p.FixedPackage != nil {
		var counter int

		for _, packages := range p.FixedPackage {
			counter += len(packages)
		}

		if counter > 0 {
			sl.ReportError(p.FixedPackage, "FixedPackage", "fixed_package", "fixed_package", "")
		}
	}
}

func (api *Api) OrderStructValidator(sl validator.StructLevel) {
	o := sl.Current().Interface().(model.OrderScalar)

	if o.PayerPhone != nil {
		num, err := libphonenumber.Parse("+380 58 4162923", "US")

		if err != nil {
			sl.ReportError(o.PayerPhone, "PayerPhone", "PayerPhone", "PayerPhone", "")
		}

		api.Order.PayerPhone = num
	}
}
