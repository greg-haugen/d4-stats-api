package platform_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/ghaugen/d4-stats-api/internal/platform"
)

func TestSentinelErrors_Distinct(t *testing.T) {
	errs := []error{
		platform.ErrPlayerNotFound,
		platform.ErrPSNTokenInvalid,
		platform.ErrPlatformRateLimited,
	}

	for i, a := range errs {
		for j, b := range errs {
			if i != j && errors.Is(a, b) {
				t.Errorf("error %d and %d are not distinct", i, j)
			}
		}
	}
}

func TestSentinelErrors_ImplementError(t *testing.T) {
	errs := []error{
		platform.ErrPlayerNotFound,
		platform.ErrPSNTokenInvalid,
		platform.ErrPlatformRateLimited,
	}
	for _, e := range errs {
		if e.Error() == "" {
			t.Errorf("error %T has empty message", e)
		}
	}
}

func TestSentinelErrors_ErrorsIs(t *testing.T) {
	wrapped := fmt.Errorf("wrapped: %w", platform.ErrPlayerNotFound)
	if !errors.Is(wrapped, platform.ErrPlayerNotFound) {
		t.Error("errors.Is should match wrapped ErrPlayerNotFound")
	}
}
