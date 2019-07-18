package api

import (
	"github.com/google/uuid"
	"github.com/ttacon/libphonenumber"
	"gopkg.in/go-playground/validator.v9"
	"regexp"
)

const (
	zipUsaRegexp = "^[0-9]{5}(?:-[0-9]{4})?$"
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
