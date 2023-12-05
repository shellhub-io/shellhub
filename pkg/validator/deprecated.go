package validator

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/shellhub-io/shellhub/pkg/models"
)

var (
	ErrInvalidFields = errors.New("invalid fields")
	ErrInvalidError  = errors.New("this error is not from a field validation")
)

var instance *validator.Validate

func init() {
	validate := validator.New()
	_ = validate.RegisterValidation("regexp", func(fl validator.FieldLevel) bool {
		_, err := regexp.Compile(fl.Field().String())

		return err == nil
	})

	_ = validate.RegisterValidation("username", func(fl validator.FieldLevel) bool {
		return regexp.MustCompile(`^([a-zA-Z0-9-_.@]){3,30}$`).MatchString(fl.Field().String())
	})

	_ = validate.RegisterValidation("device_name", func(fl validator.FieldLevel) bool {
		return regexp.MustCompile(`^([a-zA-Z0-9_-]){1,64}$`).MatchString(fl.Field().String())
	})

	instance = validate
}

func GetInstance() *validator.Validate {
	return instance
}

// getValidateTag gets the tag string from a structure field.
func getValidateTag(s interface{}, name string) (string, bool) {
	// Gets the structure's type.
	p := reflect.TypeOf(s)
	// Gets the structure's field name.
	f, ok := p.FieldByName(name)
	if !ok {
		return "", false
	}

	// Returns the "validate" fields's tag.
	return f.Tag.Get("validate"), true
}

// getInvalidFields gets the fields reported as invalids.
func getInvalidFields(err error) ([]string, error) {
	f := []string{}
	for _, err := range err.(validator.ValidationErrors) {
		f = append(f, err.Field())
	}

	return f, ErrInvalidFields
}

// GetInvalidFieldsValues receive a structure validation error and return a map with invalid fields and values.
func GetInvalidFieldsValues(err error) (map[string]interface{}, error) {
	d := make(map[string]interface{})
	for _, e := range err.(validator.ValidationErrors) {
		d[e.Field()] = e.Value()
	}

	return d, ErrInvalidFields
}

func ValidateStructFields(data interface{}) (map[string]interface{}, error) {
	if err := instance.Struct(data); err != nil {
		return GetInvalidFieldsValues(err)
	}

	return nil, nil
}

func ValidateStruct(data interface{}) ([]string, error) {
	if err := instance.Struct(data); err != nil {
		return getInvalidFields(err)
	}

	return nil, nil
}

func ValidateVar(data interface{}, tag string) ([]string, error) {
	if err := instance.Var(data, tag); err != nil {
		return getInvalidFields(err)
	}

	return nil, nil
}

// ValidateField validates if a structure's field is valid.
func ValidateField(structure interface{}, field, value string) bool {
	// Getting tag string from a structure's field.
	t, ok := getValidateTag(structure, field)
	if !ok {
		return false
	}

	// Validating the input data against the tag got.
	if _, err := ValidateVar(value, t); err != nil {
		return false
	}

	return true
}

// ValidateFieldTag validate the data for the field Tag from structure models.Device.
func ValidateFieldTag(tag string) bool {
	const Tag = "required,min=3,max=255,alphanum,ascii,excludes=/@&:"
	if _, err := ValidateVar(tag, Tag); err != nil {
		return false
	}

	return true
}

// ValidateFieldUsername validate the data for the field Username from structure models.UserData.
func ValidateFieldUsername(username string) bool {
	// Field's name that have a tag value.
	const Field = "Username"
	// Structure that contains the field above.
	s := models.UserData{}

	return ValidateField(s, Field, username)
}

// ValidateFieldEmail validate the data for the field Email from structure models.UserData.
func ValidateFieldEmail(email string) bool {
	// Field's name that have a tag value.
	const Field = "Email"
	// Structure that contains the field above.
	s := models.UserData{}

	return ValidateField(s, Field, email)
}

// ValidateFieldPassword validate the data for the field Password from structure models.UserPassword.
func ValidateFieldPassword(password string) bool {
	// Field's name that have a tag value.
	const Field = "Password"
	// Structure that contains the field above.
	s := models.UserPassword{}

	return ValidateField(s, Field, password)
}

func HashPassword(password string) string {
	s := sha256.Sum256([]byte(password))

	return hex.EncodeToString(s[:])
}

// FormatUser apply some formation rules to a models.User and encrypt the password.
func FormatUser(user *models.User) {
	user.Username = strings.ToLower(user.Username)
	user.Email = strings.ToLower(user.Email)
	if user.Password != "" {
		user.Password = HashPassword(user.Password)
	}
}

const (
	// TagRegexp is the tag used to validate a regexp.
	TagRegexp = "regexp"
	// TagUsername is the tag used to validate ShellHub's username.
	TagUsername = "username"
)
