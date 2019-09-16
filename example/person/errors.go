package person

const (
	ErrFailedValidation = Error("One or more validations failed")
)

// Error provides typed errors for this package and allow us to declare
// error constants. External packages can safely and directly determine if
// an error is one of those exported, or if the error more generally if the
// error is of the type exported by this package.
type Error string

// Error implements the standard Go Error interface
func (e Error) Error() string {
	return string(e)
}
