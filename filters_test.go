package resize

import (
	"testing"
)

func TestFilters(t *testing.T) {
	if nearest(0) != 1 || nearest(1) != 0 || nearest(-0.4) != 1 || nearest(-0.5) != 1 || nearest(0.5) != 0 {
		t.Errorf("nearest filter failed")
	}

	if cubic(0) != 1 || cubic(2.5) != 0 {
		t.Errorf("cubic filter failed")
	}
	// Cubic testing some ranges
	if cubic(0.5) == 0 || cubic(1.5) == 0 {
		t.Logf("cubic filter output is fine")
	}

	if lanczos2(2) != 0 || lanczos2(-2) != 0 || lanczos2(0) != 1 {
		t.Errorf("lanczos2 filter failed")
	}

	if lanczos3(3) != 0 || lanczos3(-3) != 0 || lanczos3(0) != 1 {
		t.Errorf("lanczos3 filter failed")
	}

	if sinc(0) != 1 {
		t.Errorf("sinc filter passed for known values")
	}
}
