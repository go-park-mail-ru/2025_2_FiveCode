package validation

import (
	"github.com/asaskevich/govalidator"
)

func init() {
	govalidator.SetFieldsRequiredByDefault(false)

	govalidator.CustomTypeTagMap.Set("password", govalidator.CustomTypeValidator(func(i interface{}, o interface{}) bool {
		s, ok := i.(string)
		if !ok {
			return false
		}
		return len(s) >= 8
	}))
}

func ValidateStruct(s interface{}) error {
	_, err := govalidator.ValidateStruct(s)
	if err != nil {
		return err
	}
	return nil
}
