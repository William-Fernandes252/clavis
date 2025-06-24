package validation

// Context provides additional information about the validation context.
type Context struct {
	Target   string                 // The target field or value being validated
	Metadata map[string]interface{} // Additional context data
}

// NewContext creates a new validation context
func NewContext(target string) Context {
	return Context{
		Target:   target,
		Metadata: make(map[string]interface{}),
	}
}

// WithMetadata adds metadata to the context
func (c Context) WithMetadata(key string, value interface{}) Context {
	// Create a new metadata map to ensure immutability
	newMetadata := make(map[string]interface{})

	// Copy existing metadata if it exists
	if c.Metadata != nil {
		for k, v := range c.Metadata {
			newMetadata[k] = v
		}
	}

	// Add the new key-value pair
	newMetadata[key] = value

	// Return a new context with the copied metadata
	return Context{
		Target:   c.Target,
		Metadata: newMetadata,
	}
}
