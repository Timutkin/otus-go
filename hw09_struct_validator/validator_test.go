package hw09structvalidator

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"
)

type UserRole string

// Test the function on different structures and other types.
type (
	User struct {
		ID     string `json:"id" validate:"len:36"`
		Name   string
		Age    int             `validate:"min:18|max:50"`
		Email  string          `validate:"regexp:^\\w+@\\w+\\.\\w+$"`
		Role   UserRole        `validate:"in:admin,stuff"`
		Phones []string        `validate:"len:11"`
		meta   json.RawMessage //nolint:unused
	}

	App struct {
		Version string `validate:"len:5"`
	}

	Token struct {
		Header    []byte
		Payload   []byte
		Signature []byte
	}

	Response struct {
		Code int    `validate:"in:200,404,500"`
		Body string `json:"omitempty"`
	}
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name        string
		in          interface{}
		expectError bool
		expectType  error
	}{
		{
			name: "valid user",
			in: User{
				ID:     "123456789012345678901234567890123456",
				Name:   "Alice",
				Age:    30,
				Email:  "alice@mail.com",
				Role:   "admin",
				Phones: []string{"12345678901"},
			},
			expectError: true,
		},
		{
			name: "invalid user - short ID, wrong email, wrong role, wrong phone len",
			in: User{
				ID:     "short",
				Name:   "Bob",
				Age:    17,
				Email:  "bob_at_mail",
				Role:   "guest",
				Phones: []string{"123"},
			},
			expectError: true,
		},
		{
			name: "invalid app - wrong len",
			in: App{
				Version: "123",
			},
			expectError: true,
		},
		{
			name: "valid app",
			in: App{
				Version: "1.0.0",
			},
			expectError: false,
		},
		{
			name: "valid response - in set",
			in: Response{
				Code: 200,
				Body: "ok",
			},
			expectError: false,
		},
		{
			name: "invalid response - code not in set",
			in: Response{
				Code: 201,
				Body: "nope",
			},
			expectError: true,
		},
		{
			name:        "not a struct",
			in:          42,
			expectError: true,
			expectType:  ErrArgumentNotStructure,
		},
		{
			name: "unsupported struct kind",
			in: Token{
				Header:    []byte("hdr"),
				Payload:   []byte("pld"),
				Signature: []byte("sig"),
			},
			expectError: true,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("case_%d_%s", i, tt.name), func(t *testing.T) {
			t.Parallel()

			err := Validate(tt.in)

			if errors.Is(tt.expectType, ErrArgumentNotStructure) {
				if !errors.Is(err, ErrArgumentNotStructure) {
					t.Errorf("expected ErrArgumentNotStructure, got %v", err)
				}
				return
			}

			if tt.expectError && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("expected no error, got: %v", err)
			}

			var errs ValidationErrors
			if tt.expectError && errors.As(err, &errs) {
				for _, e := range errs {
					if e.Field == "" || e.Err == nil {
						t.Errorf("invalid validation error: %+v", e)
					}
				}
			}
		})
	}
}
