package commands

import (
	"encoding/json"

	"github.com/William-Fernandes252/clavis/internal/errors"
	"github.com/William-Fernandes252/clavis/internal/keys"
)

// GetArgs represents the arguments for the GetCommand.
// It contains the key to retrieve the value for.
type GetArgs struct {
	Key string `json:"key" validate:"required"`
}

// GetCommand retrieves a string value by its key from the key space.
type GetCommand struct {
	kp   keys.DataKeySpace[string]
	args GetArgs
}

// NewGetCommand creates a new instance of GetCommand with the provided key space.
func NewGetCommand(kp keys.DataKeySpace[string], args GetArgs) *GetCommand {
	return &GetCommand{kp: kp, args: args}
}

// Execute runs the GetCommand and returns a generic output.
func (c *GetCommand) Execute() (*Output[string], errors.Error) {
	value, err := c.kp.Get(keys.Key(c.args.Key))
	if err != nil {
		return nil, NewCommandError("get-failed", "failed to get value", err).WithMetadata("args", c.args)
	}

	return Exit(value), nil
}

func (c *GetCommand) Name() string {
	return "get"
}

var _ Command[string] = (*GetCommand)(nil)

// GetCommandFactory creates GetCommand instances with the provided arguments.
type GetCommandFactory struct {
	kp keys.DataKeySpace[string]
}

// NewGetCommandFactory creates a new factory for GetCommand.
func NewGetCommandFactory(kp keys.DataKeySpace[string]) *GetCommandFactory {
	return &GetCommandFactory{kp: kp}
}

// Name returns the name of the command this factory creates.
func (f *GetCommandFactory) Name() string {
	return "get"
}

// Create creates a new GetCommand instance with the provided arguments.
func (f *GetCommandFactory) Create(rawArgs json.RawMessage) (Command[string], errors.Error) {
	var args GetArgs
	if err := json.Unmarshal(rawArgs, &args); err != nil {
		return nil, NewInvalidArgumentsError("failed to decode get arguments")
	}

	if err := ValidateArgs(args); err != nil {
		return nil, err
	}

	return NewGetCommand(f.kp, args), nil
}

var _ CommandFactory[string] = (*GetCommandFactory)(nil)
