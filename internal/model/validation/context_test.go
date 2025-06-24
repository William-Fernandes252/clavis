package validation

import (
	"fmt"
	"reflect"
	"testing"
)

func TestNewContext(t *testing.T) {
	type args struct {
		target string
	}
	tests := []struct {
		name string
		args args
		want Context
	}{
		{
			name: "creates context with string target",
			args: args{
				target: "username",
			},
			want: Context{
				Target:   "username",
				Metadata: make(map[string]any),
			},
		},
		{
			name: "creates context with empty target",
			args: args{
				target: "",
			},
			want: Context{
				Target:   "",
				Metadata: make(map[string]any),
			},
		},
		{
			name: "creates context with complex target path",
			args: args{
				target: "user.profile.email",
			},
			want: Context{
				Target:   "user.profile.email",
				Metadata: make(map[string]any),
			},
		},
		{
			name: "creates context with numeric target",
			args: args{
				target: "field123",
			},
			want: Context{
				Target:   "field123",
				Metadata: make(map[string]any),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewContext(tt.args.target)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewContext() = %v, want %v", got, tt.want)
			}
			// Verify metadata map is not nil
			if got.Metadata == nil {
				t.Error("NewContext() should initialize Metadata map")
			}
			// Verify metadata map is empty
			if len(got.Metadata) != 0 {
				t.Error("NewContext() should create empty Metadata map")
			}
		})
	}
}

func TestContext_WithMetadata(t *testing.T) {
	type fields struct {
		Target   string
		Metadata map[string]any
	}
	type args struct {
		key   string
		value any
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   Context
	}{
		{
			name: "adds metadata to empty context",
			fields: fields{
				Target:   "username",
				Metadata: make(map[string]any),
			},
			args: args{
				key:   "rule",
				value: "required",
			},
			want: Context{
				Target: "username",
				Metadata: map[string]any{
					"rule": "required",
				},
			},
		},
		{
			name: "adds metadata to existing context",
			fields: fields{
				Target: "password",
				Metadata: map[string]any{
					"existing": "value",
				},
			},
			args: args{
				key:   "minimum_length",
				value: 8,
			},
			want: Context{
				Target: "password",
				Metadata: map[string]any{
					"existing":       "value",
					"minimum_length": 8,
				},
			},
		},
		{
			name: "overwrites existing metadata key",
			fields: fields{
				Target: "email",
				Metadata: map[string]any{
					"format": "old_format",
				},
			},
			args: args{
				key:   "format",
				value: "email",
			},
			want: Context{
				Target: "email",
				Metadata: map[string]any{
					"format": "email",
				},
			},
		},
		{
			name: "adds different value types",
			fields: fields{
				Target:   "field",
				Metadata: make(map[string]any),
			},
			args: args{
				key:   "config",
				value: map[string]any{"nested": true, "count": 42},
			},
			want: Context{
				Target: "field",
				Metadata: map[string]any{
					"config": map[string]any{"nested": true, "count": 42},
				},
			},
		},
		{
			name: "adds nil value",
			fields: fields{
				Target:   "field",
				Metadata: make(map[string]any),
			},
			args: args{
				key:   "optional",
				value: nil,
			},
			want: Context{
				Target: "field",
				Metadata: map[string]any{
					"optional": nil,
				},
			},
		},
		{
			name: "adds metadata when metadata map is nil",
			fields: fields{
				Target:   "field",
				Metadata: nil,
			},
			args: args{
				key:   "new_key",
				value: "new_value",
			},
			want: Context{
				Target: "field",
				Metadata: map[string]any{
					"new_key": "new_value",
				},
			},
		},
		{
			name: "adds boolean value",
			fields: fields{
				Target:   "toggle",
				Metadata: make(map[string]any),
			},
			args: args{
				key:   "enabled",
				value: true,
			},
			want: Context{
				Target: "toggle",
				Metadata: map[string]any{
					"enabled": true,
				},
			},
		},
		{
			name: "adds slice value",
			fields: fields{
				Target:   "list",
				Metadata: make(map[string]any),
			},
			args: args{
				key:   "allowed_values",
				value: []string{"option1", "option2", "option3"},
			},
			want: Context{
				Target: "list",
				Metadata: map[string]any{
					"allowed_values": []string{"option1", "option2", "option3"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Context{
				Target:   tt.fields.Target,
				Metadata: tt.fields.Metadata,
			}
			got := c.WithMetadata(tt.args.key, tt.args.value)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Context.WithMetadata() = %v, want %v", got, tt.want)
			}

			// Verify the original context is unchanged (immutable pattern)
			originalHasKey := false
			if c.Metadata != nil {
				_, originalHasKey = c.Metadata[tt.args.key]
			}

			// The original should only have the key if it was already there
			expectedOriginalHasKey := false
			if tt.fields.Metadata != nil {
				_, expectedOriginalHasKey = tt.fields.Metadata[tt.args.key]
			}

			if originalHasKey != expectedOriginalHasKey {
				t.Errorf("Original context should not be modified by WithMetadata()")
			}
		})
	}
}

// Additional test for context chaining and edge cases
func TestContext_MetadataChaining(t *testing.T) {
	t.Run("chain multiple metadata additions", func(t *testing.T) {
		ctx := NewContext("user")

		result := ctx.WithMetadata("rule", "required").
			WithMetadata("min-length", 5).
			WithMetadata("max-length", 50)

		expectedMetadata := map[string]any{
			"rule":       "required",
			"min-length": 5,
			"max-length": 50,
		}

		if !reflect.DeepEqual(result.Metadata, expectedMetadata) {
			t.Errorf("Expected metadata %v, got %v", expectedMetadata, result.Metadata)
		}

		if result.Target != "user" {
			t.Errorf("Expected target 'user', got %s", result.Target)
		}

		// Verify original context is unchanged
		if len(ctx.Metadata) != 0 {
			t.Error("Original context should remain unchanged during chaining")
		}
	})

	t.Run("chaining preserves immutability", func(t *testing.T) {
		original := NewContext("field").WithMetadata("initial", "value")

		step1 := original.WithMetadata("step1", "data")
		step2 := step1.WithMetadata("step2", "more_data")

		// Each step should be independent
		if len(original.Metadata) != 1 {
			t.Errorf("Original should have 1 metadata entry, got %d", len(original.Metadata))
		}

		if len(step1.Metadata) != 2 {
			t.Errorf("Step1 should have 2 metadata entries, got %d", len(step1.Metadata))
		}

		if len(step2.Metadata) != 3 {
			t.Errorf("Step2 should have 3 metadata entries, got %d", len(step2.Metadata))
		}

		// Verify values
		if original.Metadata["initial"] != "value" {
			t.Error("Original context lost its initial value")
		}

		if step1.Metadata["step1"] != "data" {
			t.Error("Step1 context missing step1 value")
		}

		if step2.Metadata["step2"] != "more_data" {
			t.Error("Step2 context missing step2 value")
		}
	})
}

func TestContext_EdgeCases(t *testing.T) {
	t.Run("empty key", func(t *testing.T) {
		ctx := NewContext("test")
		result := ctx.WithMetadata("", "empty_key_value")

		if result.Metadata[""] != "empty_key_value" {
			t.Error("Should allow empty string as key")
		}
	})

	t.Run("unicode target and keys", func(t *testing.T) {
		ctx := NewContext("???")
		result := ctx.WithMetadata("??", "???")

		if result.Target != "???" {
			t.Errorf("Unicode target not preserved: %s", result.Target)
		}

		if result.Metadata["??"] != "???" {
			t.Error("Unicode metadata not preserved")
		}
	})

	t.Run("very long strings", func(t *testing.T) {
		longString := string(make([]byte, 10000))
		for i := range longString {
			longString = longString[:i] + "a" + longString[i+1:]
		}

		ctx := NewContext(longString)
		result := ctx.WithMetadata("long_value", longString)

		if result.Target != longString {
			t.Error("Long target string not preserved")
		}

		if result.Metadata["long_value"] != longString {
			t.Error("Long metadata value not preserved")
		}
	})

	t.Run("special characters in keys and values", func(t *testing.T) {
		specialChars := "!@#$%^&*()_+-=[]{}|;:'\",.<>?/~`"
		ctx := NewContext("field")
		result := ctx.WithMetadata(specialChars, specialChars)

		if result.Metadata[specialChars] != specialChars {
			t.Error("Special characters not preserved in metadata")
		}
	})
}

func TestContext_TypeSafety(t *testing.T) {
	t.Run("different value types", func(t *testing.T) {
		ctx := NewContext("mixed_types")

		// Test various Go types
		testCases := []struct {
			key   string
			value any
		}{
			{"string", "text"},
			{"int", 42},
			{"int64", int64(9223372036854775807)},
			{"float64", 3.14159},
			{"bool", true},
			{"slice", []int{1, 2, 3}},
			{"map", map[string]string{"nested": "value"}},
			{"nil", nil},
			{"struct", struct{ Name string }{"test"}},
			{"pointer", &[]string{"ptr_value"}},
			{"interface", any("any_value")},
		}

		result := ctx
		for _, tc := range testCases {
			result = result.WithMetadata(tc.key, tc.value)
		}

		// Verify all types are preserved
		for _, tc := range testCases {
			if !reflect.DeepEqual(result.Metadata[tc.key], tc.value) {
				t.Errorf("Type %T with key %s not preserved correctly", tc.value, tc.key)
			}
		}
	})

	t.Run("complex nested structures", func(t *testing.T) {
		type NestedStruct struct {
			Level1 map[string]any
			Level2 []any
		}

		complex := NestedStruct{
			Level1: map[string]any{
				"deep": map[string]any{
					"deeper": []int{1, 2, 3},
				},
			},
			Level2: []any{
				"string",
				42,
				map[string]bool{"flag": true},
			},
		}

		ctx := NewContext("complex")
		result := ctx.WithMetadata("complex_data", complex)

		retrieved := result.Metadata["complex_data"].(NestedStruct)
		if !reflect.DeepEqual(retrieved, complex) {
			t.Error("Complex nested structure not preserved")
		}
	})
}

func TestContext_ConcurrentSafety(t *testing.T) {
	t.Run("concurrent metadata addition", func(t *testing.T) {
		ctx := NewContext("concurrent_test")

		// Since Context.WithMetadata returns a new context,
		// it should be safe for concurrent use of the creation method
		// (though the returned contexts are separate instances)

		done := make(chan bool, 10)
		results := make([]Context, 10)

		for i := 0; i < 10; i++ {
			go func(index int) {
				defer func() { done <- true }()
				key := fmt.Sprintf("key_%d", index)
				value := fmt.Sprintf("value_%d", index)
				results[index] = ctx.WithMetadata(key, value)
			}(i)
		}

		// Wait for all goroutines
		for i := 0; i < 10; i++ {
			<-done
		}

		// Verify each result has its expected metadata
		for i, result := range results {
			expectedKey := fmt.Sprintf("key_%d", i)
			expectedValue := fmt.Sprintf("value_%d", i)

			if result.Metadata[expectedKey] != expectedValue {
				t.Errorf("Concurrent test failed for index %d", i)
			}

			// Each should have exactly one metadata entry
			if len(result.Metadata) != 1 {
				t.Errorf("Result %d should have exactly 1 metadata entry, got %d", i, len(result.Metadata))
			}
		}

		// Original context should remain unchanged
		if len(ctx.Metadata) != 0 {
			t.Error("Original context was modified during concurrent operations")
		}
	})
}

func TestContext_Integration(t *testing.T) {
	t.Run("realistic validation context usage", func(t *testing.T) {
		// Simulate a realistic validation scenario
		userCtx := NewContext("user.profile.email").
			WithMetadata("rule", "email_format").
			WithMetadata("required", true).
			WithMetadata("max-length", 254).
			WithMetadata("domain_whitelist", []string{"company.com", "partner.org"})

		// Verify the context contains all expected information
		if userCtx.Target != "user.profile.email" {
			t.Errorf("Expected target 'user.profile.email', got %s", userCtx.Target)
		}

		// Check all metadata
		expected := map[string]any{
			"rule":             "email_format",
			"required":         true,
			"max-length":       254,
			"domain_whitelist": []string{"company.com", "partner.org"},
		}

		if !reflect.DeepEqual(userCtx.Metadata, expected) {
			t.Errorf("Expected metadata %v, got %v", expected, userCtx.Metadata)
		}
	})

	t.Run("context for nested field validation", func(t *testing.T) {
		// Test validation context for deeply nested structures
		nestedCtx := NewContext("order.items[0].product.specifications.dimensions.weight").
			WithMetadata("validation_path", []string{"order", "items", "0", "product", "specifications", "dimensions", "weight"}).
			WithMetadata("parent_context", "product_validation").
			WithMetadata("validation_rules", map[string]any{
				"min":  0.1,
				"max":  1000.0,
				"unit": "kg",
			})

		if nestedCtx.Target != "order.items[0].product.specifications.dimensions.weight" {
			t.Error("Nested target path not preserved")
		}

		validationRules := nestedCtx.Metadata["validation_rules"].(map[string]any)
		if validationRules["unit"] != "kg" {
			t.Error("Nested validation rules not preserved correctly")
		}
	})
}
