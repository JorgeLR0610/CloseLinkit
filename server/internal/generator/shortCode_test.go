package generator

import (
	"regexp"
	"testing"
)

func TestNewShortCodeGenerator_InvalidLength(t *testing.T) {
	// Attempt to create a generator with a length shorter than the minimum
	_, err := NewShortCodeGenerator(2)
	if err != ErrInvalidLength {
		t.Errorf("expected error %v, got %v", ErrInvalidLength, err)
	}
}

func TestGenerateShortCode_Length(t *testing.T) {
	expectedLength := 7
	generator, _ := NewShortCodeGenerator(expectedLength)
	
	code, err := generator.GenerateShortCode()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	// Verify that the generated code has the exact expected length
	if len(code) != expectedLength {
		t.Errorf("expected length %d, got %d", expectedLength, len(code))
	}
}

func TestGenerateShortCode_Base62Characters(t *testing.T) {
	generator, _ := NewShortCodeGenerator(7)
	code, _ := generator.GenerateShortCode()
	
	// Validate that the code only contains alphanumeric Base62 characters
	matched, _ := regexp.MatchString("^[0-9a-zA-Z]+$", code)
	if !matched {
		t.Errorf("generated code contains invalid characters: %s", code)
	}
}
