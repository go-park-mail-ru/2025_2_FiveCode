package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestStruct struct {
	Email    string `valid:"email"`
	Password string `valid:"password"`
	Age      int    `valid:"range(18|100)"`
}

func TestValidateStruct(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		s := TestStruct{
			Email:    "test@example.com",
			Password: "password123",
			Age:      25,
		}
		err := ValidateStruct(s)
		assert.NoError(t, err)
	})

	t.Run("InvalidEmail", func(t *testing.T) {
		s := TestStruct{
			Email:    "invalid-email",
			Password: "password123",
			Age:      25,
		}
		err := ValidateStruct(s)
		assert.Error(t, err)
	})

	t.Run("InvalidPassword", func(t *testing.T) {
		s := TestStruct{
			Email:    "test@example.com",
			Password: "short",
			Age:      25,
		}
		err := ValidateStruct(s)
		assert.Error(t, err)
	})

	t.Run("InvalidAge", func(t *testing.T) {
		s := TestStruct{
			Email:    "test@example.com",
			Password: "password123",
			Age:      15,
		}
		err := ValidateStruct(s)
		assert.Error(t, err)
	})

	t.Run("EmptyStruct", func(t *testing.T) {
		s := TestStruct{}
		err := ValidateStruct(s)
		assert.NoError(t, err)
	})
}
