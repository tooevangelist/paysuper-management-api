package api

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/ttacon/libphonenumber"
	"gopkg.in/go-playground/validator.v9"
	"regexp"
)

const (
	zipUsaRegexp = "^[0-9]{5}(?:-[0-9]{4})?$"
	nameRegexp   = "^[\\p{L}\\p{M} \\-\\']+$"
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

func getFirstValidationError(err error) string {
	vErr := err.(validator.ValidationErrors)[0]

	return fmt.Sprintf(errorMessageMask, vErr.Field(), vErr.Tag())
}
