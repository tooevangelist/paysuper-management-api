package api

import (
	"github.com/google/uuid"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"github.com/ttacon/libphonenumber"
	"gopkg.in/go-playground/validator.v9"
	"regexp"
)

const (
	zipUsaRegexp      = "^[0-9]{5}(?:-[0-9]{4})?$"
	nameRegexp        = "^[\\p{L}\\p{M} \\-\\']+$"
	companyNameRegexp = "^[\\p{L}\\p{M} \\-\\.0-9]+$"
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
	match, err := regexp.MatchString(zipUsaRegexp, fl.Field().String())
	return match == true && err == nil
}

func (api *Api) NameValidator(fl validator.FieldLevel) bool {
	match, err := regexp.MatchString(nameRegexp, fl.Field().String())
	return match == true && err == nil
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
	match, err := regexp.MatchString(companyNameRegexp, fl.Field().String())
	return match == true && err == nil
}
