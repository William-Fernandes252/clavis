package validation

import (
	"reflect"
	"testing"

	"github.com/William-Fernandes252/clavis/internal/model/errors"
)

func TestNewValidationError(t *testing.T) {
	type args struct {
		target  string
		value   any
		message string
	}
	tests := []struct {
		name string
		args args
		want *ValidationError
	}{
		{
			name: "creates validation error with string value",
			args: args{
				target:  "username",
				value:   "test",
				message: "username is required",
			},
			want: &ValidationError{
				Target: "username",
				Value:  "test",
				ErrorData: &errors.ErrorData{
					Type:     errType,
					Code:     defaultCode,
					Message:  stringPtr("username is required"),
					Metadata: make(map[string]any),
				},
			},
		},
		{
			name: "creates validation error with numeric value",
			args: args{
				target:  "age",
				value:   -5,
				message: "age must be positive",
			},
			want: &ValidationError{
				Target: "age",
				Value:  -5,
				ErrorData: &errors.ErrorData{
					Type:     errType,
					Code:     defaultCode,
					Message:  stringPtr("age must be positive"),
					Metadata: make(map[string]any),
				},
			},
		},
		{
			name: "creates validation error with nil value",
			args: args{
				target:  "data",
				value:   nil,
				message: "data cannot be nil",
			},
			want: &ValidationError{
				Target: "data",
				Value:  nil,
				ErrorData: &errors.ErrorData{
					Type:     errType,
					Code:     defaultCode,
					Message:  stringPtr("data cannot be nil"),
					Metadata: make(map[string]any),
				},
			},
		},
		{
			name: "creates validation error with empty message",
			args: args{
				target:  "field",
				value:   "value",
				message: "",
			},
			want: &ValidationError{
				Target: "field",
				Value:  "value",
				ErrorData: &errors.ErrorData{
					Type:     errType,
					Code:     defaultCode,
					Message:  stringPtr(""),
					Metadata: make(map[string]any),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewValidationError(tt.args.target, tt.args.value, tt.args.message)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewValidationError() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Helper function for creating string pointers
func stringPtr(s string) *string {
	return &s
}

func TestValidationError_WithMetadata(t *testing.T) {
	type fields struct {
		ErrorData *errors.ErrorData
		Target    string
		Value     any
	}
	type args struct {
		key   string
		value any
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *ValidationError
	}{
		{
			name: "adds metadata to existing error",
			fields: fields{
				ErrorData: &errors.ErrorData{
					Type:     errType,
					Code:     defaultCode,
					Message:  stringPtr("test error"),
					Metadata: make(map[string]any),
				},
				Target: "username",
				Value:  "test",
			},
			args: args{
				key:   "rule",
				value: "min-length",
			},
			want: &ValidationError{
				Target: "username",
				Value:  "test",
				ErrorData: &errors.ErrorData{
					Type:     errType,
					Code:     defaultCode,
					Message:  stringPtr("test error"),
					Metadata: map[string]any{"rule": "min-length"},
				},
			},
		},
		{
			name: "adds multiple metadata entries",
			fields: fields{
				ErrorData: &errors.ErrorData{
					Type:     errType,
					Code:     defaultCode,
					Message:  stringPtr("test error"),
					Metadata: map[string]any{"existing": "value"},
				},
				Target: "age",
				Value:  10,
			},
			args: args{
				key:   "minimum",
				value: 18,
			},
			want: &ValidationError{
				Target: "age",
				Value:  10,
				ErrorData: &errors.ErrorData{
					Type:     errType,
					Code:     defaultCode,
					Message:  stringPtr("test error"),
					Metadata: map[string]any{"existing": "value", "minimum": 18},
				},
			},
		},
		{
			name: "overwrites existing metadata key",
			fields: fields{
				ErrorData: &errors.ErrorData{
					Type:     errType,
					Code:     defaultCode,
					Message:  stringPtr("test error"),
					Metadata: map[string]any{"rule": "old-value"},
				},
				Target: "email",
				Value:  "invalid@",
			},
			args: args{
				key:   "rule",
				value: "email-format",
			},
			want: &ValidationError{
				Target: "email",
				Value:  "invalid@",
				ErrorData: &errors.ErrorData{
					Type:     errType,
					Code:     defaultCode,
					Message:  stringPtr("test error"),
					Metadata: map[string]any{"rule": "email-format"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ve := &ValidationError{
				ErrorData: tt.fields.ErrorData,
				Target:    tt.fields.Target,
				Value:     tt.fields.Value,
			}
			got := ve.WithMetadata(tt.args.key, tt.args.value)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ValidationError.WithMetadata() = %v, want %v", got, tt.want)
			}
			// Verify it returns the same instance (fluent interface)
			if got != ve {
				t.Errorf("WithMetadata() should return the same instance for fluent interface")
			}
		})
	}
}

func TestValidationError_WithCode(t *testing.T) {
	type fields struct {
		ErrorData *errors.ErrorData
		Target    string
		Value     any
	}
	type args struct {
		code string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *ValidationError
	}{
		{
			name: "sets custom error code",
			fields: fields{
				ErrorData: &errors.ErrorData{
					Type:     errType,
					Code:     defaultCode,
					Message:  stringPtr("test error"),
					Metadata: make(map[string]any),
				},
				Target: "username",
				Value:  "test",
			},
			args: args{
				code: "USERNAME_TOO_SHORT",
			},
			want: &ValidationError{
				Target: "username",
				Value:  "test",
				ErrorData: &errors.ErrorData{
					Type:     errType,
					Code:     "USERNAME_TOO_SHORT",
					Message:  stringPtr("test error"),
					Metadata: make(map[string]any),
				},
			},
		},
		{
			name: "overwrites existing code",
			fields: fields{
				ErrorData: &errors.ErrorData{
					Type:     errType,
					Code:     "OLD_CODE",
					Message:  stringPtr("test error"),
					Metadata: make(map[string]any),
				},
				Target: "password",
				Value:  "weak",
			},
			args: args{
				code: "PASSWORD_TOO_WEAK",
			},
			want: &ValidationError{
				Target: "password",
				Value:  "weak",
				ErrorData: &errors.ErrorData{
					Type:     errType,
					Code:     "PASSWORD_TOO_WEAK",
					Message:  stringPtr("test error"),
					Metadata: make(map[string]any),
				},
			},
		},
		{
			name: "sets empty code",
			fields: fields{
				ErrorData: &errors.ErrorData{
					Type:     errType,
					Code:     defaultCode,
					Message:  stringPtr("test error"),
					Metadata: make(map[string]any),
				},
				Target: "field",
				Value:  "value",
			},
			args: args{
				code: "",
			},
			want: &ValidationError{
				Target: "field",
				Value:  "value",
				ErrorData: &errors.ErrorData{
					Type:     errType,
					Code:     "",
					Message:  stringPtr("test error"),
					Metadata: make(map[string]any),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ve := &ValidationError{
				ErrorData: tt.fields.ErrorData,
				Target:    tt.fields.Target,
				Value:     tt.fields.Value,
			}
			got := ve.WithCode(tt.args.code)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ValidationError.WithCode() = %v, want %v", got, tt.want)
			}
			// Verify it returns the same instance (fluent interface)
			if got != ve {
				t.Errorf("WithCode() should return the same instance for fluent interface")
			}
		})
	}
}

func TestValidationError_Error(t *testing.T) {
	type fields struct {
		ErrorData *errors.ErrorData
		Target    string
		Value     any
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "returns custom message when present",
			fields: fields{
				ErrorData: &errors.ErrorData{
					Type:     errType,
					Code:     defaultCode,
					Message:  stringPtr("Custom validation message"),
					Metadata: make(map[string]any),
				},
				Target: "username",
				Value:  "test",
			},
			want: "Custom validation message",
		},
		{
			name: "returns default message when no custom message",
			fields: fields{
				ErrorData: &errors.ErrorData{
					Type:     errType,
					Code:     defaultCode,
					Message:  nil,
					Metadata: make(map[string]any),
				},
				Target: "email",
				Value:  "invalid@email",
			},
			want: "email: validation failed for \"invalid@email\"",
		},
		{
			name: "handles nil value in default message",
			fields: fields{
				ErrorData: &errors.ErrorData{
					Type:     errType,
					Code:     defaultCode,
					Message:  nil,
					Metadata: make(map[string]any),
				},
				Target: "data",
				Value:  nil,
			},
			want: "data: validation failed for \"<nil>\"",
		},
		{
			name: "handles numeric value in default message",
			fields: fields{
				ErrorData: &errors.ErrorData{
					Type:     errType,
					Code:     defaultCode,
					Message:  nil,
					Metadata: make(map[string]any),
				},
				Target: "age",
				Value:  -5,
			},
			want: "age: validation failed for \"-5\"",
		},
		{
			name: "returns empty string when message is empty",
			fields: fields{
				ErrorData: &errors.ErrorData{
					Type:     errType,
					Code:     defaultCode,
					Message:  stringPtr(""),
					Metadata: make(map[string]any),
				},
				Target: "field",
				Value:  "value",
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ve := ValidationError{
				ErrorData: tt.fields.ErrorData,
				Target:    tt.fields.Target,
				Value:     tt.fields.Value,
			}
			if got := ve.Error(); got != tt.want {
				t.Errorf("ValidationError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidationError_Type(t *testing.T) {
	type fields struct {
		ErrorData *errors.ErrorData
		Target    string
		Value     any
	}
	tests := []struct {
		name   string
		fields fields
		want   errors.ErrorType
	}{
		{
			name: "returns validation error type",
			fields: fields{
				ErrorData: &errors.ErrorData{
					Type:     errType,
					Code:     defaultCode,
					Message:  stringPtr("test error"),
					Metadata: make(map[string]any),
				},
				Target: "username",
				Value:  "test",
			},
			want: errType,
		},
		{
			name: "returns correct type even with different error data",
			fields: fields{
				ErrorData: &errors.ErrorData{
					Type:     "custom-type",
					Code:     "CUSTOM_CODE",
					Message:  stringPtr("custom message"),
					Metadata: map[string]any{"key": "value"},
				},
				Target: "field",
				Value:  123,
			},
			want: "custom-type",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ve := ValidationError{
				ErrorData: tt.fields.ErrorData,
				Target:    tt.fields.Target,
				Value:     tt.fields.Value,
			}
			if got := ve.Type(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ValidationError.Type() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidationError_Code(t *testing.T) {
	type fields struct {
		ErrorData *errors.ErrorData
		Target    string
		Value     any
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "returns default code",
			fields: fields{
				ErrorData: &errors.ErrorData{
					Type:     errType,
					Code:     defaultCode,
					Message:  stringPtr("test error"),
					Metadata: make(map[string]any),
				},
				Target: "username",
				Value:  "test",
			},
			want: defaultCode,
		},
		{
			name: "returns custom code",
			fields: fields{
				ErrorData: &errors.ErrorData{
					Type:     errType,
					Code:     "CUSTOM_ERROR_CODE",
					Message:  stringPtr("test error"),
					Metadata: make(map[string]any),
				},
				Target: "email",
				Value:  "invalid@",
			},
			want: "CUSTOM_ERROR_CODE",
		},
		{
			name: "returns empty code",
			fields: fields{
				ErrorData: &errors.ErrorData{
					Type:     errType,
					Code:     "",
					Message:  stringPtr("test error"),
					Metadata: make(map[string]any),
				},
				Target: "field",
				Value:  "value",
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ve := ValidationError{
				ErrorData: tt.fields.ErrorData,
				Target:    tt.fields.Target,
				Value:     tt.fields.Value,
			}
			if got := ve.Code(); got != tt.want {
				t.Errorf("ValidationError.Code() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidationError_Metadata(t *testing.T) {
	type fields struct {
		ErrorData *errors.ErrorData
		Target    string
		Value     any
	}
	tests := []struct {
		name   string
		fields fields
		want   map[string]any
	}{
		{
			name: "returns empty metadata",
			fields: fields{
				ErrorData: &errors.ErrorData{
					Type:     errType,
					Code:     defaultCode,
					Message:  stringPtr("test error"),
					Metadata: make(map[string]any),
				},
				Target: "username",
				Value:  "test",
			},
			want: make(map[string]any),
		},
		{
			name: "returns populated metadata",
			fields: fields{
				ErrorData: &errors.ErrorData{
					Type:    errType,
					Code:    defaultCode,
					Message: stringPtr("test error"),
					Metadata: map[string]any{
						"rule":    "min-length",
						"minimum": 5,
						"actual":  3,
					},
				},
				Target: "password",
				Value:  "abc",
			},
			want: map[string]any{
				"rule":    "min-length",
				"minimum": 5,
				"actual":  3,
			},
		},
		{
			name: "returns nil metadata when nil",
			fields: fields{
				ErrorData: &errors.ErrorData{
					Type:     errType,
					Code:     defaultCode,
					Message:  stringPtr("test error"),
					Metadata: nil,
				},
				Target: "field",
				Value:  "value",
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ve := ValidationError{
				ErrorData: tt.fields.ErrorData,
				Target:    tt.fields.Target,
				Value:     tt.fields.Value,
			}
			if got := ve.Metadata(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ValidationError.Metadata() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewValidationResult(t *testing.T) {
	tests := []struct {
		name string
		want *ValidationResult
	}{
		{
			name: "creates empty validation result",
			want: &ValidationResult{
				Errors: make([]ValidationError, 0),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewValidationResult()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewValidationResult() = %v, want %v", got, tt.want)
			}
			// Verify the slice is not nil
			if got.Errors == nil {
				t.Errorf("NewValidationResult() should initialize Errors slice")
			}
			// Verify it starts empty
			if len(got.Errors) != 0 {
				t.Errorf("NewValidationResult() should start with empty Errors slice")
			}
		})
	}
}

func TestValidationResult_Error(t *testing.T) {
	type fields struct {
		Errors []ValidationError
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "returns default message for empty errors",
			fields: fields{
				Errors: []ValidationError{},
			},
			want: "validation failed",
		},
		{
			name: "returns single error message",
			fields: fields{
				Errors: []ValidationError{
					{
						Target: "username",
						Value:  "test",
						ErrorData: &errors.ErrorData{
							Type:     errType,
							Code:     defaultCode,
							Message:  stringPtr("username is required"),
							Metadata: make(map[string]any),
						},
					},
				},
			},
			want: "username is required",
		},
		{
			name: "joins multiple error messages with semicolon",
			fields: fields{
				Errors: []ValidationError{
					{
						Target: "username",
						Value:  "",
						ErrorData: &errors.ErrorData{
							Type:     errType,
							Code:     defaultCode,
							Message:  stringPtr("username is required"),
							Metadata: make(map[string]any),
						},
					},
					{
						Target: "password",
						Value:  "123",
						ErrorData: &errors.ErrorData{
							Type:     errType,
							Code:     defaultCode,
							Message:  stringPtr("password too short"),
							Metadata: make(map[string]any),
						},
					},
				},
			},
			want: "username is required; password too short",
		},
		{
			name: "handles errors with default messages",
			fields: fields{
				Errors: []ValidationError{
					{
						Target: "email",
						Value:  "invalid@",
						ErrorData: &errors.ErrorData{
							Type:     errType,
							Code:     defaultCode,
							Message:  nil,
							Metadata: make(map[string]any),
						},
					},
					{
						Target: "age",
						Value:  -5,
						ErrorData: &errors.ErrorData{
							Type:     errType,
							Code:     defaultCode,
							Message:  stringPtr("age must be positive"),
							Metadata: make(map[string]any),
						},
					},
				},
			},
			want: "email: validation failed for \"invalid@\"; age must be positive",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ve := ValidationResult{
				Errors: tt.fields.Errors,
			}
			if got := ve.Error(); got != tt.want {
				t.Errorf("ValidationResult.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidationResult_ToJSON(t *testing.T) {
	type fields struct {
		Errors []ValidationError
	}
	tests := []struct {
		name    string
		fields  fields
		want    []byte
		wantErr bool
	}{
		{
			name: "serializes empty validation result",
			fields: fields{
				Errors: []ValidationError{},
			},
			want:    []byte(`{"errors":[]}`),
			wantErr: false,
		},
		{
			name: "serializes single validation error",
			fields: fields{
				Errors: []ValidationError{
					{
						Target: "username",
						Value:  "test",
						ErrorData: &errors.ErrorData{
							Type:     errType,
							Code:     defaultCode,
							Message:  stringPtr("username is required"),
							Metadata: make(map[string]any),
						},
					},
				},
			},
			want:    []byte(`{"errors":[{"type":"validation","code":"validation-failed","message":"username is required","target":"username","value":"test"}]}`),
			wantErr: false,
		},
		{
			name: "serializes multiple validation errors",
			fields: fields{
				Errors: []ValidationError{
					{
						Target: "username",
						Value:  "",
						ErrorData: &errors.ErrorData{
							Type:     errType,
							Code:     "USERNAME_REQUIRED",
							Message:  stringPtr("username is required"),
							Metadata: map[string]any{"rule": "required"},
						},
					},
					{
						Target: "password",
						Value:  "123",
						ErrorData: &errors.ErrorData{
							Type:     errType,
							Code:     "PASSWORD_TOO_SHORT",
							Message:  stringPtr("password too short"),
							Metadata: map[string]any{"minimum": 8, "actual": 3},
						},
					},
				},
			},
			want:    []byte(`{"errors":[{"type":"validation","code":"USERNAME_REQUIRED","metadata":{"rule":"required"},"message":"username is required","target":"username","value":""},{"type":"validation","code":"PASSWORD_TOO_SHORT","metadata":{"actual":3,"minimum":8},"message":"password too short","target":"password","value":"123"}]}`),
			wantErr: false,
		},
		{
			name: "serializes validation error with nil value",
			fields: fields{
				Errors: []ValidationError{
					{
						Target: "data",
						Value:  nil,
						ErrorData: &errors.ErrorData{
							Type:     errType,
							Code:     defaultCode,
							Message:  stringPtr("data cannot be nil"),
							Metadata: make(map[string]any),
						},
					},
				},
			},
			want:    []byte(`{"errors":[{"type":"validation","code":"validation-failed","message":"data cannot be nil","target":"data","value":null}]}`),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ve := ValidationResult{
				Errors: tt.fields.Errors,
			}
			got, err := ve.ToJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidationResult.ToJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ValidationResult.ToJSON() = %s, want %s", string(got), string(tt.want))
			}
		})
	}
}

func TestValidationResult_HasErrors(t *testing.T) {
	type fields struct {
		Errors []ValidationError
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "returns false for empty errors",
			fields: fields{
				Errors: []ValidationError{},
			},
			want: false,
		},
		{
			name: "returns false for nil errors",
			fields: fields{
				Errors: nil,
			},
			want: false,
		},
		{
			name: "returns true for single error",
			fields: fields{
				Errors: []ValidationError{
					{
						Target: "username",
						Value:  "test",
						ErrorData: &errors.ErrorData{
							Type:     errType,
							Code:     defaultCode,
							Message:  stringPtr("username is required"),
							Metadata: make(map[string]any),
						},
					},
				},
			},
			want: true,
		},
		{
			name: "returns true for multiple errors",
			fields: fields{
				Errors: []ValidationError{
					{
						Target: "username",
						Value:  "",
						ErrorData: &errors.ErrorData{
							Type:     errType,
							Code:     defaultCode,
							Message:  stringPtr("username is required"),
							Metadata: make(map[string]any),
						},
					},
					{
						Target: "password",
						Value:  "123",
						ErrorData: &errors.ErrorData{
							Type:     errType,
							Code:     defaultCode,
							Message:  stringPtr("password too short"),
							Metadata: make(map[string]any),
						},
					},
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ve := ValidationResult{
				Errors: tt.fields.Errors,
			}
			if got := ve.HasErrors(); got != tt.want {
				t.Errorf("ValidationResult.HasErrors() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidationResult_Add(t *testing.T) {
	type fields struct {
		Errors []ValidationError
	}
	type args struct {
		err ValidationError
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		verify func(*testing.T, *ValidationResult)
	}{
		{
			name: "adds error to empty result",
			fields: fields{
				Errors: []ValidationError{},
			},
			args: args{
				err: ValidationError{
					Target: "username",
					Value:  "test",
					ErrorData: &errors.ErrorData{
						Type:     errType,
						Code:     defaultCode,
						Message:  stringPtr("username is required"),
						Metadata: make(map[string]any),
					},
				},
			},
			verify: func(t *testing.T, vr *ValidationResult) {
				if len(vr.Errors) != 1 {
					t.Errorf("Expected 1 error, got %d", len(vr.Errors))
				}
				if vr.Errors[0].Target != "username" {
					t.Errorf("Expected target 'username', got '%s'", vr.Errors[0].Target)
				}
			},
		},
		{
			name: "adds error to existing errors",
			fields: fields{
				Errors: []ValidationError{
					{
						Target: "existing",
						Value:  "value",
						ErrorData: &errors.ErrorData{
							Type:     errType,
							Code:     defaultCode,
							Message:  stringPtr("existing error"),
							Metadata: make(map[string]any),
						},
					},
				},
			},
			args: args{
				err: ValidationError{
					Target: "new",
					Value:  "value",
					ErrorData: &errors.ErrorData{
						Type:     errType,
						Code:     "NEW_ERROR",
						Message:  stringPtr("new error"),
						Metadata: make(map[string]any),
					},
				},
			},
			verify: func(t *testing.T, vr *ValidationResult) {
				if len(vr.Errors) != 2 {
					t.Errorf("Expected 2 errors, got %d", len(vr.Errors))
				}
				if vr.Errors[0].Target != "existing" {
					t.Errorf("Expected first error target 'existing', got '%s'", vr.Errors[0].Target)
				}
				if vr.Errors[1].Target != "new" {
					t.Errorf("Expected second error target 'new', got '%s'", vr.Errors[1].Target)
				}
				if vr.Errors[1].Code() != "NEW_ERROR" {
					t.Errorf("Expected second error code 'NEW_ERROR', got '%s'", vr.Errors[1].Code())
				}
			},
		},
		{
			name: "preserves order of errors",
			fields: fields{
				Errors: []ValidationError{},
			},
			args: args{
				err: ValidationError{
					Target: "first",
					Value:  "value",
					ErrorData: &errors.ErrorData{
						Type:     errType,
						Code:     "FIRST",
						Message:  stringPtr("first error"),
						Metadata: make(map[string]any),
					},
				},
			},
			verify: func(t *testing.T, vr *ValidationResult) {
				// Add a second error
				secondErr := ValidationError{
					Target: "second",
					Value:  "value",
					ErrorData: &errors.ErrorData{
						Type:     errType,
						Code:     "SECOND",
						Message:  stringPtr("second error"),
						Metadata: make(map[string]any),
					},
				}
				vr.Add(secondErr)

				if len(vr.Errors) != 2 {
					t.Errorf("Expected 2 errors, got %d", len(vr.Errors))
				}
				if vr.Errors[0].Code() != "FIRST" {
					t.Errorf("Expected first error code 'FIRST', got '%s'", vr.Errors[0].Code())
				}
				if vr.Errors[1].Code() != "SECOND" {
					t.Errorf("Expected second error code 'SECOND', got '%s'", vr.Errors[1].Code())
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ve := &ValidationResult{
				Errors: tt.fields.Errors,
			}
			ve.Add(tt.args.err)
			tt.verify(t, ve)
		})
	}
}

// Additional test for ValidationError ToJSON method
func TestValidationError_ToJSON(t *testing.T) {
	tests := []struct {
		name    string
		ve      ValidationError
		want    []byte
		wantErr bool
	}{
		{
			name: "serializes validation error with all fields",
			ve: ValidationError{
				Target: "username",
				Value:  "test",
				ErrorData: &errors.ErrorData{
					Type:     errType,
					Code:     "USERNAME_REQUIRED",
					Message:  stringPtr("username is required"),
					Metadata: map[string]any{"rule": "required"},
				},
			},
			want:    []byte(`{"type":"validation","code":"USERNAME_REQUIRED","metadata":{"rule":"required"},"message":"username is required","target":"username","value":"test"}`),
			wantErr: false,
		},
		{
			name: "serializes validation error with nil value",
			ve: ValidationError{
				Target: "data",
				Value:  nil,
				ErrorData: &errors.ErrorData{
					Type:     errType,
					Code:     defaultCode,
					Message:  stringPtr("data cannot be nil"),
					Metadata: make(map[string]any),
				},
			},
			want:    []byte(`{"type":"validation","code":"validation-failed","message":"data cannot be nil","target":"data","value":null}`),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.ve.ToJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidationError.ToJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ValidationError.ToJSON() = %s, want %s", string(got), string(tt.want))
			}
		})
	}
}

// Test for interface compliance
func TestValidationError_InterfaceCompliance(t *testing.T) {
	ve := NewValidationError("test", "value", "test message")

	// Test that it implements error interface
	var _ error = ve

	// Test that it implements errors.Error interface
	var _ errors.Error = ve

	// Test all required methods are callable
	if ve.Error() == "" {
		t.Error("Error() should return non-empty string")
	}
	if ve.Type() != errType {
		t.Errorf("Type() should return %s", errType)
	}
	if ve.Code() != defaultCode {
		t.Errorf("Code() should return %s", defaultCode)
	}
	if ve.Metadata() == nil {
		t.Error("Metadata() should not return nil")
	}
	if _, err := ve.ToJSON(); err != nil {
		t.Errorf("ToJSON() should not error: %v", err)
	}
}

// Test fluent interface chaining
func TestValidationError_FluentInterface(t *testing.T) {
	ve := NewValidationError("test", "value", "test message")

	// Test chaining methods
	result := ve.WithCode("CUSTOM_CODE").WithMetadata("key1", "value1").WithMetadata("key2", 42)

	// Verify it's the same instance
	if result != ve {
		t.Error("Fluent methods should return the same instance")
	}

	// Verify changes were applied
	if ve.Code() != "CUSTOM_CODE" {
		t.Errorf("Expected code 'CUSTOM_CODE', got '%s'", ve.Code())
	}

	metadata := ve.Metadata()
	if metadata["key1"] != "value1" {
		t.Errorf("Expected metadata key1='value1', got '%v'", metadata["key1"])
	}
	if metadata["key2"] != 42 {
		t.Errorf("Expected metadata key2=42, got '%v'", metadata["key2"])
	}
}
