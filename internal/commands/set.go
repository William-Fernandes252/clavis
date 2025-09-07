package commands

import (
	"encoding/json"

	"github.com/William-Fernandes252/clavis/internal/errors"
	"github.com/William-Fernandes252/clavis/internal/keys"
)

// SetArgs represents the arguments for the SetCommand.
// It contains the key and value to set.
type SetArgs struct {
	Key            string `json:"key" validate:"required"`
	Value          string `json:"value" validate:"required"`
	Overwrite      bool   `json:"overwrite,omitempty"`                   // If true, overwrite existing value
	Expiration     int64  `json:"expiration,omitempty" validate:"min=0"` // Optional expiration time in unix timestamp format
	KeepExpiration bool   `json:"keep_expiration,omitempty"`             // If true, keep the existing expiration time if the key already exists
	Get            bool   `json:"get,omitempty"`                         // If true, return the existing value if the key already exists
}

// SetCommand sets a string value by its key in the key space.
type SetCommand struct {
	kp   keys.DataKeySpace[string]
	em   keys.ExpirationManager
	args SetArgs
}

// NewSetCommand creates a new instance of SetCommand with the provided key space.
func NewSetCommand(kp keys.DataKeySpace[string], em keys.ExpirationManager, args SetArgs) *SetCommand {
	return &SetCommand{kp: kp, em: em, args: args}
}

// Execute runs the SetCommand and returns a generic output.
func (c *SetCommand) Execute() (*Output[*string], errors.Error) {
	k := keys.Key(c.args.Key)

	// Check if key exists and capture previous value (for Get and overwrite logic)
	existed := false
	var prevVal string
	if v, err := c.kp.Get(k); err != nil {
		if err.Code() != keys.KeyNotFoundCode {
			return nil, NewCommandError("get-existing-failed", "failed to check existing value", err).WithMetadata("args", c.args)
		}
	} else {
		existed = true
		prevVal = v
	}

	// Write the new value only if not existing or overwrite requested
	if !(existed && !c.args.Overwrite) {
		if err := c.kp.Set(k, c.args.Value); err != nil {
			return nil, NewCommandError("set-failed", "failed to set value", err).WithMetadata("args", c.args)
		}
	}

	// Handle expiration behavior independently of write
	if c.em != nil {
		if c.args.KeepExpiration {
			if !existed {
				// No TTL to keep; if a new expiration is provided for a new key, schedule it.
				if c.args.Expiration > 0 {
					if err := c.em.Schedule(k, keys.FromUnix(c.args.Expiration)); err != nil {
						return nil, NewCommandError("schedule-expiration-failed", "failed to schedule expiration", err).WithMetadata("args", c.args)
					}
				}
			}
			// If existed, KeepExpiration means do nothing (preserve current TTL)
		} else {
			if c.args.Expiration > 0 {
				if err := c.em.Schedule(k, keys.FromUnix(c.args.Expiration)); err != nil {
					return nil, NewCommandError("schedule-expiration-failed", "failed to schedule expiration", err).WithMetadata("args", c.args)
				}
			} else {
				// No expiration provided => clear any existing expiration
				c.em.Cancel(k)
			}
		}
	}

	// Prepare optional previous value in response if requested
	var outPrev *string
	if c.args.Get && existed {
		p := prevVal
		outPrev = &p
	}

	return Exit(outPrev), nil
}

func (c *SetCommand) Name() string {
	return "set"
}

var _ Command[*string] = (*SetCommand)(nil)

// SetCommandFactory creates SetCommand instances with the provided arguments.
type SetCommandFactory struct {
	kp keys.DataKeySpace[string]
	em keys.ExpirationManager
}

// NewSetCommandFactory creates a new factory for SetCommand.
func NewSetCommandFactory(kp keys.DataKeySpace[string], em keys.ExpirationManager) *SetCommandFactory {
	return &SetCommandFactory{kp: kp, em: em}
}

// Name returns the name of the command this factory creates.
func (f *SetCommandFactory) Name() string {
	return "set"
}

// Create creates a new SetCommand instance with the provided arguments.
func (f *SetCommandFactory) Create(rawArgs json.RawMessage) (Command[*string], errors.Error) {
	var args SetArgs
	if err := json.Unmarshal(rawArgs, &args); err != nil {
		return nil, NewInvalidArgumentsError("failed to decode set arguments")
	}

	if err := ValidateArgs(args); err != nil {
		return nil, err
	}

	return NewSetCommand(f.kp, f.em, args), nil
}

// Ensure factories implement the CommandFactory interface
var _ CommandFactory[*string] = (*SetCommandFactory)(nil)
