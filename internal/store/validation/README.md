# Validation Store Implementation

This document describes the validation wrapper that adds data validation functionality to any `Store` implementation.

## Overview

The `ValidatedStore` is a decorator/wrapper that adds validation capabilities to any existing `Store` implementation. It uses functional composition to provide flexible, composable validation rules for both keys and values.

## Features

- **Decorator pattern**: Wraps any existing Store implementation
- **Functional validation**: Uses function composition for flexible validation rules
- **Built-in validators**: Common validation functions for typical use cases
- **Composable**: Chain multiple validation functions together
- **Type-safe**: Compile-time validation of validator functions
- **Error handling**: Clear, descriptive error messages

## Architecture

```
???????????????????    validates    ???????????????????
? ValidatedStore  ? ??????????????? ? Underlying Store?
?                 ?                 ? (Memory/Badger) ?
???????????????????                 ???????????????????
```

The `ValidatedStore` intercepts all operations and applies validation before delegating to the underlying store.

## Usage

### Basic Usage with Default Validators

```go
// Create underlying store (Memory or BadgerDB)
baseStore, err := memory.NewWithDefaults()
if err != nil {
    log.Fatal(err)
}

// Wrap with default validation
validatedStore := validation.NewWithDefaultValidators(baseStore)
defer validatedStore.Close()

// Operations are now validated
err = validatedStore.Put("valid-key", []byte("valid-value")) // ? Succeeds
err = validatedStore.Put("", []byte("value"))                // ? Fails: empty key
```

### Custom Validators

```go
// Create custom key validator
keyValidator := validation.ComposeKeyValidators(
    validation.ValidateNonEmptyKey,
    validation.ValidateKeyLength(256),
    func(key string) error {
        if !strings.HasPrefix(key, "user:") {
            return fmt.Errorf("key must start with 'user:' prefix")
        }
        return nil
    },
)

// Create custom value validator
valueValidator := validation.ComposeValueValidators(
    validation.ValidateValueSize(1024*1024), // 1MB limit
    func(key string, value []byte) error {
        if !json.Valid(value) {
            return fmt.Errorf("value must be valid JSON")
        }
        return nil
    },
)

// Create validated store with custom validators
validatedStore := validation.New(baseStore, keyValidator, valueValidator)
```

## Built-in Validators

### Key Validators

#### ValidateNonEmptyKey
Ensures keys are not empty strings.

```go
var ValidateNonEmptyKey KeyValidator = func(key string) error {
    if key == "" {
        return fmt.Errorf("key cannot be empty")
    }
    return nil
}
```

**Usage:**
```go
// Direct usage
err := ValidateNonEmptyKey("") // Returns: "key cannot be empty"
err := ValidateNonEmptyKey("valid-key") // Returns: nil
```

#### ValidateKeyLength
Creates a validator that enforces maximum key length.

```go
func ValidateKeyLength(maxLength int) KeyValidator
```

**Parameters:**
- `maxLength`: Maximum allowed key length in characters

**Usage:**
```go
validator := ValidateKeyLength(100)
err := validator("short-key")                    // Returns: nil
err := validator(strings.Repeat("a", 200))       // Returns: error
```

### Value Validators

#### ValidateValueSize
Creates a validator that enforces maximum value size in bytes.

```go
func ValidateValueSize(maxSize int64) ValueValidator
```

**Parameters:**
- `maxSize`: Maximum allowed value size in bytes

**Usage:**
```go
validator := ValidateValueSize(1024) // 1KB limit
err := validator("key", []byte("small"))           // Returns: nil
err := validator("key", make([]byte, 2048))        // Returns: error
```

## Validator Composition

### ComposeKeyValidators
Combines multiple key validators into a single validator function.

```go
func ComposeKeyValidators(validators ...KeyValidator) KeyValidator
```

**Usage:**
```go
composedValidator := ComposeKeyValidators(
    ValidateNonEmptyKey,
    ValidateKeyLength(100),
    func(key string) error {
        if strings.Contains(key, " ") {
            return fmt.Errorf("key cannot contain spaces")
        }
        return nil
    },
)
```

### ComposeValueValidators
Combines multiple value validators into a single validator function.

```go
func ComposeValueValidators(validators ...ValueValidator) ValueValidator
```

**Usage:**
```go
composedValidator := ComposeValueValidators(
    ValidateValueSize(1024*1024), // 1MB
    func(key string, value []byte) error {
        if len(value) == 0 {
            return fmt.Errorf("value cannot be empty")
        }
        return nil
    },
)
```

## Default Validation Configuration

The `NewWithDefaultValidators` function creates a store with sensible default validation rules:

```go
func NewWithDefaultValidators(s store.Store) *ValidatedStore {
    keyValidator := ComposeKeyValidators(
        ValidateNonEmptyKey,
        ValidateKeyLength(1024), // 1KB key limit
    )

    valueValidator := ComposeValueValidators(
        ValidateValueSize(100 * 1024 * 1024), // 100MB value limit
    )

    return New(s, keyValidator, valueValidator)
}
```

**Default Limits:**
- **Maximum key length**: 1,024 characters
- **Maximum value size**: 100 MB
- **Key requirements**: Non-empty

## Custom Validator Examples

### Domain-Specific Validators

```go
// User ID validation
func ValidateUserKey(key string) error {
    if !strings.HasPrefix(key, "user:") {
        return fmt.Errorf("user keys must start with 'user:' prefix")
    }

    userID := strings.TrimPrefix(key, "user:")
    if len(userID) < 3 {
        return fmt.Errorf("user ID must be at least 3 characters")
    }

    return nil
}

// Email validation for user values
func ValidateEmailValue(key string, value []byte) error {
    email := string(value)
    if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
        return fmt.Errorf("value must be a valid email address")
    }
    return nil
}

// Usage
keyValidator := ComposeKeyValidators(
    ValidateNonEmptyKey,
    ValidateUserKey,
)

valueValidator := ComposeValueValidators(
    ValidateValueSize(256), // Email shouldn't be too long
    ValidateEmailValue,
)

store := validation.New(baseStore, keyValidator, valueValidator)
```

### JSON Validation

```go
func ValidateJSONValue(key string, value []byte) error {
    if !json.Valid(value) {
        return fmt.Errorf("value must be valid JSON")
    }
    return nil
}

// Usage
valueValidator := ComposeValueValidators(
    ValidateValueSize(10*1024*1024), // 10MB limit for JSON
    ValidateJSONValue,
)
```

### Pattern Matching

```go
func ValidateKeyPattern(pattern *regexp.Regexp) KeyValidator {
    return func(key string) error {
        if !pattern.MatchString(key) {
            return fmt.Errorf("key does not match required pattern: %s", pattern.String())
        }
        return nil
    }
}

// Usage
pattern := regexp.MustCompile(`^[a-z]+:[0-9]+$`) // prefix:number format
keyValidator := ComposeKeyValidators(
    ValidateNonEmptyKey,
    ValidateKeyPattern(pattern),
)
```

## API Reference

### Types

```go
type KeyValidator func(string) error
type ValueValidator func(string, []byte) error
```

### Constructor Functions

- `New(s store.Store, keyValidator KeyValidator, valueValidator ValueValidator) *ValidatedStore`
- `NewWithDefaultValidators(s store.Store) *ValidatedStore`

### Built-in Validators

- `ValidateNonEmptyKey` - Rejects empty keys
- `ValidateKeyLength(maxLength int) KeyValidator` - Enforces key length limits
- `ValidateValueSize(maxSize int64) ValueValidator` - Enforces value size limits

### Composition Functions

- `ComposeKeyValidators(validators ...KeyValidator) KeyValidator`
- `ComposeValueValidators(validators ...ValueValidator) ValueValidator`

### Store Interface Methods

The `ValidatedStore` implements the complete `Store` interface:

- `Get(key string) ([]byte, bool, error)` - Validates key before retrieval
- `Put(key string, value []byte) error` - Validates key and value before storage
- `Delete(key string) error` - Validates key before deletion
- `Scan(prefix string) (map[string][]byte, error)` - Passes through to underlying store
- `Close() error` - Closes underlying store

## Error Handling

Validation errors are returned immediately without calling the underlying store:

```go
err := store.Put("", []byte("value"))
// Returns: "key cannot be empty"

err := store.Put("valid-key", make([]byte, 200*1024*1024))
// Returns: "value too large: maximum 104857600 bytes, got 209715200"
```

## Performance Considerations

- **Validation overhead**: Minimal for simple validators, may increase with complex regex or JSON validation
- **Early failure**: Failed validation prevents unnecessary calls to underlying store
- **Memory usage**: Validators operate on provided data without copying

## Testing

The validation package includes comprehensive tests:

```bash
go test ./internal/store/validation/... -v
```

Test coverage includes:
- Default validator behavior
- Custom validator composition
- Error message accuracy
- Integration with different store implementations

## Integration Examples

### With MemoryStore

```go
memStore, _ := memory.NewWithDefaults()
validatedStore := validation.NewWithDefaultValidators(memStore)
```

### With BadgerStore

```go
badgerStore, _ := badger.NewWithPath("/tmp/test-db")
validatedStore := validation.NewWithDefaultValidators(badgerStore)
```

### Chain Multiple Validation Layers

```go
// Base store
baseStore, _ := memory.NewWithDefaults()

// First validation layer: basic validation
basicValidated := validation.NewWithDefaultValidators(baseStore)

// Second validation layer: domain-specific validation
userValidator := func(key string) error {
    if !strings.HasPrefix(key, "user:") {
        return fmt.Errorf("only user keys allowed")
    }
    return nil
}

domainValidated := validation.New(
    basicValidated,
    userValidator,
    func(key string, value []byte) error { return nil }, // No additional value validation
)
```

## Best Practices

1. **Compose validators**: Use composition functions for multiple validation rules
2. **Fail fast**: Place cheaper validations first in composition chains
3. **Clear error messages**: Provide descriptive error messages for failed validations
4. **Test validators**: Unit test custom validators independently
5. **Document requirements**: Clearly document validation requirements for your application
6. **Consider performance**: Complex validations may impact performance on high-volume operations

## Use Cases

The validation store is ideal for:

- **API backends**: Validate user input before storage
- **Data integrity**: Ensure data meets application requirements
- **Multi-tenant applications**: Enforce tenant-specific key patterns
- **Compliance**: Enforce regulatory data requirements
- **Development/testing**: Catch data issues early in development
