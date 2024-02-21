package bc

// Validator returns an error from the Validate method
type Validator interface {
	Validate() error
}
