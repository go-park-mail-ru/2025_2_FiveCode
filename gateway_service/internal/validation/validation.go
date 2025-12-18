package validation

import (
	"regexp"

	"github.com/asaskevich/govalidator"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z0-9]{2,}$`)

func init() {
	govalidator.SetFieldsRequiredByDefault(false)

	govalidator.CustomTypeTagMap.Set("password", govalidator.CustomTypeValidator(func(i interface{}, o interface{}) bool {
		s, ok := i.(string)
		if !ok {
			return false
		}
		return len(s) >= 8
	}))

	govalidator.TagMap["email"] = govalidator.Validator(func(str string) bool {
		return emailRegex.MatchString(str)
	})
}

func ValidateStruct(s interface{}) error {
	_, err := govalidator.ValidateStruct(s)
	if err != nil {
		return err
	}
	return nil
}
