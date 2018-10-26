package api

import (
	"github.com/ProtocolONE/p1pay.api/database/model"
	"gopkg.in/go-playground/validator.v9"
)

func ProjectStructValidator(sl validator.StructLevel) {
	p := sl.Current().Interface().(model.ProjectScalar)

	if p.SendNotifyEmail == true && len(p.NotifyEmails) <= 0 {
		sl.ReportError(p.NotifyEmails, "NotifyEmails", "notify_emails", "notify_emails", "")
	}
}
