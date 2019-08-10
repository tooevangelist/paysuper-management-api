package api

import (
	"github.com/google/uuid"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"github.com/ttacon/libphonenumber"
	"gopkg.in/go-playground/validator.v9"
	"regexp"
)

var (
	availablePositions = map[string]bool{
		UserProfilePositionCEO:               true,
		UserProfilePositionCTO:               true,
		UserProfilePositionCMO:               true,
		UserProfilePositionCFO:               true,
		UserProfilePositionProjectManagement: true,
		UserProfilePositionGenericManagement: true,
		UserProfilePositionSoftwareDeveloper: true,
		UserProfilePositionMarketing:         true,
		UserProfilePositionSupport:           true,
	}

	availableAnnualIncome = []*grpc.RangeInt{
		{From: 0, To: 1000},
		{From: 1000, To: 10000},
		{From: 10000, To: 100000},
		{From: 100000, To: 1000000},
		{From: 1000000, To: 0},
	}

	availableNumberOfEmployees = []*grpc.RangeInt{
		{From: 1, To: 10},
		{From: 11, To: 50},
		{From: 51, To: 100},
		{From: 100, To: 0},
	}

	zipUsaRegexp      = regexp.MustCompile("^[0-9]{5}(?:-[0-9]{4})?$")
	nameRegexp        = regexp.MustCompile("^[\\p{L}\\p{M} \\-\\']+$")
	companyNameRegexp = regexp.MustCompile("^[\\p{L}\\p{M} \\-\\.0-9]+$")
	zipGeneralRegexp  = regexp.MustCompile("^\\d{0,30}$")
	swiftRegexp       = regexp.MustCompile("^[A-Z]{6}[A-Z0-9]{2}([A-Z0-9]{3})?$")
	cityRegexp        = regexp.MustCompile("^[\\p{L}\\p{M} \\-\\.]+$")
)

func (api *Api) PhoneValidator(fl validator.FieldLevel) bool {
	_, err := libphonenumber.Parse(fl.Field().String(), "US")
	return err == nil
}

func (api *Api) UuidValidator(fl validator.FieldLevel) bool {
	_, err := uuid.Parse(fl.Field().String())
	return err == nil
}

func (api *Api) ZipUsaValidator(fl validator.FieldLevel) bool {
	return zipUsaRegexp.MatchString(fl.Field().String())
}

func (api *Api) NameValidator(fl validator.FieldLevel) bool {
	return nameRegexp.MatchString(fl.Field().String())
}

func (api *Api) PositionValidator(fl validator.FieldLevel) bool {
	_, ok := availablePositions[fl.Field().String()]
	return ok
}

func (api *Api) CompanyValidator(sl validator.StructLevel) {
	company := sl.Current().Interface().(grpc.UserProfileCompany)
	res := api.rangeIntValidator(company.AnnualIncome, availableAnnualIncome)

	if res == false {
		sl.ReportError(company.AnnualIncome, "AnnualIncome", "annual_income", "annual_income", "")
	}

	res = api.rangeIntValidator(company.NumberOfEmployees, availableNumberOfEmployees)

	if res == false {
		sl.ReportError(company.NumberOfEmployees, "NumberOfEmployees", "number_of_employees", "number_of_employees", "")
	}
}

func (api *Api) rangeIntValidator(in *grpc.RangeInt, rng []*grpc.RangeInt) bool {
	for _, v := range rng {
		if in.From == v.From && in.To == v.To {
			return true
		}
	}

	return false
}

func (api *Api) CompanyNameValidator(fl validator.FieldLevel) bool {
	return companyNameRegexp.MatchString(fl.Field().String())
}

func (api *Api) MerchantCompanyValidator(sl validator.StructLevel) {
	company := sl.Current().Interface().(billing.MerchantCompanyInfo)

	reg := zipGeneralRegexp

	if v, ok := zipRegexp[company.Country]; ok {
		reg = v
	}

	match := reg.MatchString(company.Zip)

	if !match {
		sl.ReportError(company.Zip, "Zip", "zip", "zip", "")
	}
}

func (api *Api) SwiftValidator(fl validator.FieldLevel) bool {
	return swiftRegexp.MatchString(fl.Field().String())
}

func (api *Api) CityValidator(fl validator.FieldLevel) bool {
	return cityRegexp.MatchString(fl.Field().String())
}
