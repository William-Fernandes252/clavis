package errors

import (
	"reflect"
	"testing"
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
				ErrorData: &ErrorData{
					Type:     errType,
					Code:     defaultCode,
					Message:  "username is required",
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
				ErrorData: &ErrorData{
					Type:     errType,
					Code:     defaultCode,
					Message:  "age must be positive",
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
				ErrorData: &ErrorData{
					Type:     errType,
					Code:     defaultCode,
					Message:  "data cannot be nil",
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
				ErrorData: &ErrorData{
					Type:     errType,
					Code:     defaultCode,
					Message:  "",
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

func TestValidationError_WithMetadata(t *testing.T) {
	type fields struct {
		ErrorData *ErrorData
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
				ErrorData: &ErrorData{
					Type:     errType,
					Code:     defaultCode,
					Message:  "test error",
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
				ErrorData: &ErrorData{
					Type:     errType,
					Code:     defaultCode,
					Message:  "test error",
					Metadata: map[string]any{"rule": "min-length"},
				},
			},
		},
		{
			name: "adds multiple metadata entries",
			fields: fields{
				ErrorData: &ErrorData{
					Type:     errType,
					Code:     defaultCode,
					Message:  "test error",
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
				ErrorData: &ErrorData{
					Type:     errType,
					Code:     defaultCode,
					Message:  "test error",
					Metadata: map[string]any{"existing": "value", "minimum": 18},
				},
			},
		},
		{
			name: "overwrites existing metadata key",
			fields: fields{
				ErrorData: &ErrorData{
					Type:     errType,
					Code:     defaultCode,
					Message:  "test error",
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
				ErrorData: &ErrorData{
					Type:     errType,
					Code:     defaultCode,
					Message:  "test error",
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
		ErrorData *ErrorData
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
				ErrorData: &ErrorData{
					Type:     errType,
					Code:     defaultCode,
					Message:  "test error",
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
				ErrorData: &ErrorData{
					Type:     errType,
					Code:     "USERNAME_TOO_SHORT",
					Message:  "test error",
					Metadata: make(map[string]any),
				},
			},
		},
		{
			name: "overwrites existing code",
			fields: fields{
				ErrorData: &ErrorData{
					Type:     errType,
					Code:     "OLD_CODE",
					Message:  "test error",
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
				ErrorData: &ErrorData{
					Type:     errType,
					Code:     "PASSWORD_TOO_WEAK",
					Message:  "test error",
					Metadata: make(map[string]any),
				},
			},
		},
		{
			name: "sets empty code",
			fields: fields{
				ErrorData: &ErrorData{
					Type:     errType,
					Code:     defaultCode,
					Message:  "test error",
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
				ErrorData: &ErrorData{
					Type:     errType,
					Code:     "",
					Message:  "test error",
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
		ErrorData *ErrorData
		Target    string
		Value     any
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "returns target and custom message",
			fields: fields{
				ErrorData: &ErrorData{
					Type:     errType,
					Code:     defaultCode,
					Message:  "invalid username",
					Metadata: make(map[string]any),
				},
				Target: "username",
				Value:  "test",
			},
			want: "username: invalid username",
		},
		{
			name: "returns empty string when message is empty",
			fields: fields{
				ErrorData: &ErrorData{
					Type:     errType,
					Code:     defaultCode,
					Message:  "",
					Metadata: make(map[string]any),
				},
				Target: "field",
				Value:  "value",
			},
			want: "field: ",
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
		ErrorData *ErrorData
		Target    string
		Value     any
	}
	tests := []struct {
		name   string
		fields fields
		want   ErrorType
	}{
		{
			name: "returns validation error type",
			fields: fields{
				ErrorData: &ErrorData{
					Type:     errType,
					Code:     defaultCode,
					Message:  "test error",
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
				ErrorData: &ErrorData{
					Type:     "custom-type",
					Code:     "CUSTOM_CODE",
					Message:  "custom message",
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
		ErrorData *ErrorData
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
				ErrorData: &ErrorData{
					Type:     errType,
					Code:     defaultCode,
					Message:  "test error",
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
				ErrorData: &ErrorData{
					Type:     errType,
					Code:     "CUSTOM_ERROR_CODE",
					Message:  "test error",
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
				ErrorData: &ErrorData{
					Type:     errType,
					Code:     "",
					Message:  "test error",
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
		ErrorData *ErrorData
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
				ErrorData: &ErrorData{
					Type:     errType,
					Code:     defaultCode,
					Message:  "test error",
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
				ErrorData: &ErrorData{
					Type:    errType,
					Code:    defaultCode,
					Message: "test error",
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
				ErrorData: &ErrorData{
					Type:     errType,
					Code:     defaultCode,
					Message:  "test error",
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
