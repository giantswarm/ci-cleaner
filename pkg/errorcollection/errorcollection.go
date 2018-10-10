// Package errorcollection provides a error type to store multiple errors.
package errorcollection

import (
	"fmt"
)

// ErrorCollection is our error type.
type ErrorCollection struct {
	errors []error
}

// Error is the function we need to implement the error interface.
func (ec *ErrorCollection) Error() string {
	return fmt.Sprintf("collection of %d errors", len(ec.errors))
}

// Append adds an error to the collection.
func (ec *ErrorCollection) Append(e error) {
	ec.errors = append(ec.errors, e)
}

// HasErrors returns true if the collection has errors in it.
func (ec *ErrorCollection) HasErrors() bool {
	if len(ec.errors) > 0 {
		return true
	}
	return false
}

// Dump returns printable string of all contained errors.
func (ec *ErrorCollection) Dump() string {
	if !ec.HasErrors() {
		return "No errors."
	}

	s := ""
	for _, e := range ec.errors {
		if innerEC, ok := e.(*ErrorCollection); ok {
			s += innerEC.Dump()
		} else {
			s += fmt.Sprintf("- %s\n", e)
		}
	}
	return s
}

// Errors returns all stored errors
func (ec *ErrorCollection) Errors() []error {
	return ec.errors
}
