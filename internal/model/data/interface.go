package data

import (
	"github.com/William-Fernandes252/clavis/internal/model/errors"
	"github.com/William-Fernandes252/clavis/internal/model/validation"
	"github.com/William-Fernandes252/clavis/internal/model/validation/validators"
)

// Data is the core interface every data type must implement.
// T is the in‐memory Go representation (e.g. string, map[string]string, []string, etc.)
type Data[T any] interface {
    // Type returns the type name ("string", "hash", "list", …)
    Type() string

    // Validators returns a slice of validators to run on any new value
    Validators() []validators.Validator[T]

    // Validate runs all validators, returning the first error (or nil)
    Validate(value T) validation.ValidationError

    // Parse takes the raw command arguments (strings) and builds a T
    //
    // e.g. for strings: args = ["myvalue"] → "myvalue"
    //      for hashes:  args = ["field1","v1","field2","v2"] → map[field1:v1 field2:v2]
    Parse(args []string) (T, errors.Error)

    // Serialize turns a T into bytes suitable for storage in your key–value store
    //
    // e.g. []byte("hello"), JSON‐marshal for maps, gob for structs, etc.
    Serialize(value T) ([]byte, errors.Error)

    // Deserialize reads the bytes back into a T
    Deserialize(raw []byte) (T, errors.Error)

    // Format turns a T into a string or slice of strings for client replies
    //
    // e.g. for lists: []string{"item1","item2"}, for hashes: ["field1","v1",…]
    Format(value T) ([]string, errors.Error)
}
