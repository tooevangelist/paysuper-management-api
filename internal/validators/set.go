package validators

import (
	"context"
	"github.com/ProtocolONE/go-core/v2/pkg/logger"
	"github.com/ProtocolONE/go-core/v2/pkg/provider"
	"github.com/go-pascal/iban"
	"github.com/google/uuid"
	billPkg "github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"github.com/paysuper/paysuper-management-api/internal/dispatcher/common"
	"github.com/ttacon/libphonenumber"
	"gopkg.in/go-playground/validator.v9"
	"regexp"
)

type ValidatorSet struct {
	services common.Services
	provider.LMT
}

var (
	availablePositions = map[string]bool{
		common.UserProfilePositionCEO:               true,
		common.UserProfilePositionCTO:               true,
		common.UserProfilePositionCMO:               true,
		common.UserProfilePositionCFO:               true,
		common.UserProfilePositionProjectManagement: true,
		common.UserProfilePositionGenericManagement: true,
		common.UserProfilePositionSoftwareDeveloper: true,
		common.UserProfilePositionMarketing:         true,
		common.UserProfilePositionSupport:           true,
	}

	availableAnnualIncome = []*billing.RangeInt{
		{From: 0, To: 1000},
		{From: 1000, To: 10000},
		{From: 10000, To: 100000},
		{From: 100000, To: 1000000},
		{From: 1000000, To: 0},
	}

	availableNumberOfEmployees = []*billing.RangeInt{
		{From: 1, To: 10},
		{From: 11, To: 50},
		{From: 51, To: 100},
		{From: 100, To: 0},
	}

	zipUsaRegexp      = regexp.MustCompile("^[0-9]{5}(?:-[0-9]{4})?$")
	nameRegexp        = regexp.MustCompile("^[\\p{L}\\p{M} \\-\\']+$")
	companyNameRegexp = regexp.MustCompile("^[\\p{L}\\p{M} \\-\\.0-9\"]+$")
	zipGeneralRegexp  = regexp.MustCompile("^\\d{0,30}$")
	swiftRegexp       = regexp.MustCompile("^[A-Z]{6}[A-Z0-9]{2}([A-Z0-9]{3})?$")
	cityRegexp        = regexp.MustCompile("^[\\p{L}\\p{M} \\-\\.]+$")
	localeRegexp      = regexp.MustCompile("^[a-z]{2}-[A-Z]{2,10}$")
)

// ProductPriceValidator
func (v *ValidatorSet) ProductPriceValidator(fl validator.FieldLevel) bool {
	value := fl.Field().Interface()
	prices, ok := value.([]*billing.ProductPrice)
	if !ok {
		price, ok := value.(*billing.ProductPrice)
		if !ok {
			return false
		}
		prices = append(prices, price)
	}

	for _, price := range prices {
		if price.IsVirtualCurrency == true {
			continue
		}

		if len(price.Currency) == 0 || len(price.Region) == 0 {
			return false
		}
	}

	return true
}

// PhoneValidator
func (v *ValidatorSet) PhoneValidator(fl validator.FieldLevel) bool {
	_, err := libphonenumber.Parse(fl.Field().String(), "US")
	return err == nil
}

// PriceRegionValidator validates group price region for existing in dictionary
func (v *ValidatorSet) PriceRegionValidator(fl validator.FieldLevel) bool {
	region := fl.Field().String()
	if region == "" {
		return true
	}

	resp, err := v.services.Billing.GetPriceGroupByRegion(context.TODO(), &grpc.GetPriceGroupByRegionRequest{Region: region})
	if err != nil {
		v.L().Error("can't get price region", logger.PairArgs("method", "PriceRegionValidator"),
			logger.PairArgs("region", region),
			logger.PairArgs("err", err))
		return false
	}

	return resp.Group != nil
}

// UuidValidator
func (v *ValidatorSet) UuidValidator(fl validator.FieldLevel) bool {
	_, err := uuid.Parse(fl.Field().String())
	return err == nil
}

// ZipUsaValidator
func (v *ValidatorSet) ZipUsaValidator(fl validator.FieldLevel) bool {
	return zipUsaRegexp.MatchString(fl.Field().String())
}

// NameValidator
func (v *ValidatorSet) NameValidator(fl validator.FieldLevel) bool {
	return nameRegexp.MatchString(fl.Field().String())
}

// PositionValidator
func (v *ValidatorSet) PositionValidator(fl validator.FieldLevel) bool {
	_, ok := availablePositions[fl.Field().String()]
	return ok
}

// CompanyValidator
func (v *ValidatorSet) CompanyValidator(sl validator.StructLevel) {
	company := sl.Current().Interface().(grpc.UserProfileCompany)
	res := v.RangeIntValidator(company.AnnualIncome, availableAnnualIncome)

	if res == false {
		sl.ReportError(company.AnnualIncome, "AnnualIncome", "annual_income", "annual_income", "")
	}

	res = v.RangeIntValidator(company.NumberOfEmployees, availableNumberOfEmployees)

	if res == false {
		sl.ReportError(company.NumberOfEmployees, "NumberOfEmployees", "number_of_employees", "number_of_employees", "")
	}
}

// RangeIntValidator
func (v *ValidatorSet) RangeIntValidator(in *billing.RangeInt, rng []*billing.RangeInt) bool {
	for _, v := range rng {
		if in.From == v.From && in.To == v.To {
			return true
		}
	}

	return false
}

// CompanyNameValidator
func (v *ValidatorSet) CompanyNameValidator(fl validator.FieldLevel) bool {
	return companyNameRegexp.MatchString(fl.Field().String())
}

// MerchantCompanyValidator
func (v *ValidatorSet) MerchantCompanyValidator(sl validator.StructLevel) {
	company := sl.Current().Interface().(billing.MerchantCompanyInfo)

	reg := zipGeneralRegexp

	if v, ok := common.ZipRegexp[company.Country]; ok {
		reg = v
	}

	match := reg.MatchString(company.Zip)

	if !match {
		sl.ReportError(company.Zip, "Zip", "zip", "zip", "")
	}
}

// SwiftValidator
func (v *ValidatorSet) SwiftValidator(fl validator.FieldLevel) bool {
	return swiftRegexp.MatchString(fl.Field().String())
}

// CityValidator
func (v *ValidatorSet) CityValidator(fl validator.FieldLevel) bool {
	return cityRegexp.MatchString(fl.Field().String())
}

// WorldRegionValidator
func (v *ValidatorSet) WorldRegionValidator(fl validator.FieldLevel) bool {
	_, ok := common.TariffRegions[fl.Field().String()]
	return ok
}

// TariffRegionValidator
func (v *ValidatorSet) TariffRegionValidator(fl validator.FieldLevel) bool {
	_, ok := billPkg.HomeRegions[fl.Field().String()]
	return ok
}

// IBAN validator
func (v *ValidatorSet) IBANValidator(fl validator.FieldLevel) bool {
	_, err := iban.NewIBAN(fl.Field().String())
	return err == nil
}

// User locale validator
func (v *ValidatorSet) UserLocaleValidator(fl validator.FieldLevel) bool {
	return localeRegexp.MatchString(fl.Field().String())
}

// New
func New(services common.Services, set provider.AwareSet) *ValidatorSet {
	set.Logger = set.Logger.WithFields(logger.Fields{"service": Prefix})
	return &ValidatorSet{services: services, LMT: &set}
}
