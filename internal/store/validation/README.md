# Validation Store Implementation

This document describes the validation wrapper that adds data validation functionality to any `Store` implementation using the new validation framework.

## Overview

The `ValidatedStore` is a decorator/wrapper that adds validation capabilities to any existing `Store` implementation. It leverages the robust validation framework from `internal/model/validation` to provide flexible, composable validation rules for both keys and values.

## Features

- **Decorator pattern**: Wraps any existing Store implementation
- **Built-in validators**: Pre-configured validators for common use cases
- **Composable**: Chain multiple validators using ValidatorChain
- **Type-safe**: Strongly typed validators with compile-time safety
- **Rich error handling**: Detailed error messages with metadata and error codes
- **Helper functions**: Convenient validator factories for common patterns

## Architecture

```
┌─────────────────┐    validates    ┌─────────────────┐
│ ValidatedStore  │ ──────────────► │ Underlying Store│
│                 │                 │ (Memory/Badger) │
│ ┌─────────────┐ │                 └─────────────────┘
│ │ StoreKey    │ │
│ │ Validator   │ │
│ └─────────────┘ │
│ ┌─────────────┐ │
│ │ StoreValue  │ │
│ │ Validator   │ │
│ └─────────────┘ │
└─────────────────┘
```

The `ValidatedStore` uses `StoreKeyValidator` and `StoreValueValidator` objects that wrap the validation framework's validators and provide compatibility with the store interface.

## Usage

### Basic Usage with Default Validators

```go
import (
    "github.com/William-Fernandes252/clavis/internal/store/memory"
    "github.com/William-Fernandes252/clavis/internal/store/validation"
)

// Create underlying store (Memory or BadgerDB)
baseStore, err := memory.NewWithDefaults()
if err != nil {
    log.Fatal(err)
}

// Wrap with default validation
validatedStore := validation.NewWithDefaultValidators(baseStore)
defer validatedStore.Close()

// Operations are now validated
err = validatedStore.Put("valid-key", []byte("valid-value")) // ✓ Succeeds
err = validatedStore.Put("", []byte("value"))                // ✗ Fails: empty key
err = validatedStore.Put("key", make([]byte, 200*1024*1024)) // ✗ Fails: value too large
```

### Custom Validators

```go
import (
    "github.com/William-Fernandes252/clavis/internal/model/validation/validators"
    "github.com/William-Fernandes252/clavis/internal/store/validation"
)

// Create custom key validator with multiple rules
keyValidator := validation.NewStoreKeyValidator(
    validation.NonEmptyKeyValidator(),
    validation.KeyLengthValidator(256),
    validators.Custom(func(key string) bool {
        return strings.HasPrefix(key, "user:")
    }, "key must start with 'user:' prefix").WithName("user-prefix"),
)

// Create custom value validator
valueValidator := validation.NewStoreValueValidator(
    validation.ValueSizeValidator(1024*1024), // 1MB limit
    validation.ValueContentValidator(func(value []byte) bool {
        return json.Valid(value)
    }, "value must be valid JSON"),
)

// Create validated store with custom validators
validatedStore := validation.New(baseStore, keyValidator, valueValidator)
defer validatedStore.Close()
```

## Validator Types

### StoreKeyValidator

The `StoreKeyValidator` wraps a chain of validators from the validation framework:

```go
type StoreKeyValidator struct {
    chain *validators.ValidatorChain[string]
}

func NewStoreKeyValidator(validatorList ...validators.Validator[string]) *StoreKeyValidator
func (skv *StoreKeyValidator) Validate(key string) error
```

**Key Features:**

- Uses `ValidatorChain` for composing multiple validators
- Automatically converts validation results to traditional errors
- Supports all validators from the validation framework

### StoreValueValidator

The `StoreValueValidator` handles value validation with access to both key and value:

```go
type StoreValueValidator struct {
    validators []func(key string, value []byte, ctx validation.Context) *validation.ValidationError
}

func NewStoreValueValidator(validatorFuncs ...func(string, []byte, validation.Context) *validation.ValidationError) *StoreValueValidator
func (svv *StoreValueValidator) Validate(key string, value []byte) error
```

**Key Features:**

- Access to both key and value during validation
- Rich validation context with metadata
- Supports custom validation functions

## Built-in Validator Helpers

### Key Validators

#### NonEmptyKeyValidator

Ensures keys are not empty strings.

```go
func NonEmptyKeyValidator() validators.Validator[string]
```

**Usage:**

```go
keyValidator := validation.NewStoreKeyValidator(
    validation.NonEmptyKeyValidator(),
)
```

#### KeyLengthValidator

Creates a validator that enforces maximum key length.

```go
func KeyLengthValidator(maxLength int) validators.Validator[string]
```

**Usage:**

```go
keyValidator := validation.NewStoreKeyValidator(
    validation.KeyLengthValidator(100), // Max 100 characters
)
```

#### KeyPatternValidator

Creates a validator that enforces a regex pattern.

```go
func KeyPatternValidator(pattern string) validators.Validator[string]
```

**Usage:**

```go
keyValidator := validation.NewStoreKeyValidator(
    validation.KeyPatternValidator(`^user:[a-z0-9]+$`),
)
```

### Value Validators

#### ValueSizeValidator

Creates a validator that enforces maximum value size in bytes.

```go
func ValueSizeValidator(maxSize int64) func(string, []byte, validation.Context) *validation.ValidationError
```

**Usage:**

```go
valueValidator := validation.NewStoreValueValidator(
    validation.ValueSizeValidator(1024*1024), // 1MB limit
)
```

#### ValueContentValidator

Creates a validator for custom content validation.

```go
func ValueContentValidator(validateFn func([]byte) bool, errorMsg string) func(string, []byte, validation.Context) *validation.ValidationError
```

**Usage:**

```go
valueValidator := validation.NewStoreValueValidator(
    validation.ValueContentValidator(func(value []byte) bool {
        return json.Valid(value)
    }, "value must be valid JSON"),
)
```

## Default Validation Configuration

The `NewWithDefaultValidators` function creates a store with sensible default validation rules:

```go
func NewWithDefaultValidators(s store.Store) *ValidatedStore {
    keyValidator := NewStoreKeyValidator(DefaultKeyValidators()...)
    valueValidator := NewStoreValueValidator(DefaultValueValidators()...)
    return New(s, keyValidator, valueValidator)
}
```

**Default Rules:**

- **Key validation**: Non-empty, maximum 1024 characters
- **Value validation**: Maximum 100MB size

**Default Helper Functions:**

```go
func DefaultKeyValidators() []validators.Validator[string]
func DefaultValueValidators() []func(string, []byte, validation.Context) *validation.ValidationError
```

## Advanced Examples

### Domain-Specific User Validation

```go
// User key validation with custom rules
userKeyValidator := validation.NewStoreKeyValidator(
    validation.NonEmptyKeyValidator(),
    validators.Custom(func(key string) bool {
        if !strings.HasPrefix(key, "user:") {
            return false
        }
        parts := strings.Split(key, ":")
        if len(parts) != 2 {
            return false
        }
        userId := parts[1]
        return len(userId) >= 3 // Minimum user ID length
    }, "invalid user key format").WithName("user-key-format"),
)

// User value validation requiring email field
userValueValidator := validation.NewStoreValueValidator(
    validation.ValueSizeValidator(1024), // 1KB max for user data
    validation.ValueContentValidator(func(value []byte) bool {
        return strings.Contains(string(value), "email")
    }, "user data must contain email field"),
)

userStore := validation.New(baseStore, userKeyValidator, userValueValidator)
```

### Composite Validation with Multiple Rules

```go
// Complex key validation
keyValidator := validation.NewStoreKeyValidator(
    validation.NonEmptyKeyValidator(),
    validation.KeyLengthValidator(50),
    validators.Custom(func(key string) bool {
        return !strings.Contains(key, "banned")
    }, "key cannot contain 'banned'").WithName("no-banned"),
    validators.Custom(func(key string) bool {
        return strings.Contains(key, ":")
    }, "key must contain ':'").WithName("requires-colon"),
)

// Complex value validation
valueValidator := validation.NewStoreValueValidator(
    validation.ValueSizeValidator(100),
    validation.ValueContentValidator(func(value []byte) bool {
        return !strings.Contains(string(value), "forbidden")
    }, "value cannot contain 'forbidden'"),
)

compositeStore := validation.New(baseStore, keyValidator, valueValidator)
```

### JSON Schema Validation

```go
import "encoding/json"

// JSON value validator
jsonValueValidator := validation.NewStoreValueValidator(
    validation.ValueSizeValidator(10*1024*1024), // 10MB limit for JSON
    validation.ValueContentValidator(func(value []byte) bool {
        return json.Valid(value)
    }, "value must be valid JSON"),
    validation.ValueContentValidator(func(value []byte) bool {
        var data map[string]interface{}
        if err := json.Unmarshal(value, &data); err != nil {
            return false
        }
        // Require specific fields
        _, hasID := data["id"]
        _, hasName := data["name"]
        return hasID && hasName
    }, "JSON must contain 'id' and 'name' fields"),
)
```

## API Reference

### Constructor Functions

```go
// Create store with custom validators
func New(s store.Store, keyValidator *StoreKeyValidator, valueValidator *StoreValueValidator) *ValidatedStore

// Create store with default validators
func NewWithDefaultValidators(s store.Store) *ValidatedStore
```

### Validator Constructors

```go
// Key validator constructor
func NewStoreKeyValidator(validatorList ...validators.Validator[string]) *StoreKeyValidator

// Value validator constructor
func NewStoreValueValidator(validatorFuncs ...func(string, []byte, validation.Context) *validation.ValidationError) *StoreValueValidator
```

### Built-in Validator Helpers

**Key Validators:**

```go
func NonEmptyKeyValidator() validators.Validator[string]
func KeyLengthValidator(maxLength int) validators.Validator[string]
func KeyPatternValidator(pattern string) validators.Validator[string]
```

**Value Validators:**

```go
func ValueSizeValidator(maxSize int64) func(string, []byte, validation.Context) *validation.ValidationError
func ValueContentValidator(validateFn func([]byte) bool, errorMsg string) func(string, []byte, validation.Context) *validation.ValidationError
```

**Default Validators:**

```go
func DefaultKeyValidators() []validators.Validator[string]
func DefaultValueValidators() []func(string, []byte, validation.Context) *validation.ValidationError
```

### Store Interface Methods

The `ValidatedStore` implements the complete `Store` interface:

- `Get(key string) ([]byte, bool, error)` - Validates key before retrieval
- `Put(key string, value []byte) error` - Validates key and value before storage
- `Delete(key string) error` - Validates key before deletion
- `Scan(prefix string) (map[string][]byte, error)` - Validates prefix, returns all matching data
- `Close() error` - Closes underlying store

## Error Handling

The validation system provides detailed error messages with context:

```go
// Empty key error
err := store.Put("", []byte("value"))
// Returns: "key cannot be empty"

// Value too large error
err := store.Put("valid-key", make([]byte, 200*1024*1024))
// Returns: "value too large: maximum 104857600 bytes, got 209715200"

// Custom validation error
err := store.Put("invalid:format", []byte("value"))
// Returns: "key must start with 'user:' prefix"
```

## Performance Considerations

- **Validation overhead**: Modern validation framework is optimized for performance
- **Early failure**: Failed validation prevents unnecessary calls to underlying store
- **Memory efficiency**: Validators operate without copying data
- **Chain optimization**: Validator chains are pre-compiled for efficiency

## Testing

The validation package includes comprehensive tests with helper functions:

```bash
go test ./internal/store/validation/... -v
```

**Test coverage includes:**

- Default validator behavior
- Custom validator composition
- Error message accuracy and formatting
- Integration with different store implementations
- Performance benchmarks

## Integration Examples

### With MemoryStore

```go
import (
    "github.com/William-Fernandes252/clavis/internal/store/memory"
    "github.com/William-Fernandes252/clavis/internal/store/validation"
)

memStore, _ := memory.NewWithDefaults()
validatedStore := validation.NewWithDefaultValidators(memStore)
defer validatedStore.Close()
```

### With BadgerStore

```go
import (
    "github.com/William-Fernandes252/clavis/internal/store/badger"
    "github.com/William-Fernandes252/clavis/internal/store/validation"
)

badgerStore, _ := badger.NewWithPath("/tmp/test-db")
validatedStore := validation.NewWithDefaultValidators(badgerStore)
defer validatedStore.Close()
```

### Layered Validation

```go
// Base store
baseStore, _ := memory.NewWithDefaults()

// First validation layer: basic validation
basicValidated := validation.NewWithDefaultValidators(baseStore)

// Second validation layer: domain-specific validation
domainKeyValidator := validation.NewStoreKeyValidator(
    validators.Custom(func(key string) bool {
        return strings.HasPrefix(key, "app:")
    }, "keys must start with 'app:' prefix").WithName("app-prefix"),
)

domainValueValidator := validation.NewStoreValueValidator(
    // No additional value validation
)

domainValidated := validation.New(basicValidated, domainKeyValidator, domainValueValidator)
defer domainValidated.Close()
```

## Best Practices

1. **Use default validators**: Start with `NewWithDefaultValidators()` for basic protection
2. **Compose validators**: Combine multiple validation rules using the validator constructors
3. **Order matters**: Place cheaper validations first in validator chains
4. **Clear error messages**: Provide descriptive error messages in custom validators
5. **Test validators**: Unit test custom validators independently
6. **Use helper functions**: Leverage the built-in validator helpers for common patterns
7. **Consider performance**: Complex validations may impact high-volume operations
8. **Document requirements**: Clearly document validation requirements for your application

## Migration from Legacy API

If you're migrating from the old function-based validation API:

**Old API:**

```go
// Old way
keyValidator := validation.ComposeKeyValidators(
    validation.ValidateNonEmptyKey,
    validation.ValidateKeyLength(100),
)

valueValidator := validation.ComposeValueValidators(
    validation.ValidateValueSize(1024),
)

store := validation.New(baseStore, keyValidator, valueValidator)
```

**New API:**

```go
// New way
keyValidator := validation.NewStoreKeyValidator(
    validation.NonEmptyKeyValidator(),
    validation.KeyLengthValidator(100),
)

valueValidator := validation.NewStoreValueValidator(
    validation.ValueSizeValidator(1024),
)

store := validation.New(baseStore, keyValidator, valueValidator)
```

## Use Cases

The validation store is ideal for:

- **API backends**: Validate user input before storage
- **Data integrity**: Ensure data meets application requirements
- **Multi-tenant applications**: Enforce tenant-specific key patterns
- **Compliance**: Enforce regulatory data requirements
- **Development/testing**: Catch data issues early in development
- **Microservices**: Consistent validation across service boundaries
- **Data pipelines**: Validate data at ingestion points
