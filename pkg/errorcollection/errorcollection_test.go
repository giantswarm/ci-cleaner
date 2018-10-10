package errorcollection

import (
	"errors"
	"testing"

	"github.com/giantswarm/microerror"
)

func TestErrorCollection(t *testing.T) {
	ec := &ErrorCollection{}

	if ec.HasErrors() {
		t.Error("HasErrors should return false here, but returns true")
	}

	ec.Append(microerror.Mask(errors.New("this is the first error")))
	ec.Append(microerror.Mask(errors.New("this is the second error")))

	if !ec.HasErrors() {
		t.Error("HasErrors should return true here, but returns false")
	}

	expectedOutput := "collection of 2 errors"

	if ec.Error() != expectedOutput {
		t.Errorf("expected %q, got %q", expectedOutput, ec.Error())
	}
}
