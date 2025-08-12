package commands

import (
	"github.com/William-Fernandes252/clavis/internal/errors"
	"github.com/William-Fernandes252/clavis/internal/keys"
)

// GetCommand retrieves a string value by its key from the key space.
type GetCommand struct {
	kp keys.DataKeySpace[string]
}

// NewGetCommand creates a new instance of GetCommand with the provided key space.
func NewGetCommand(kp keys.DataKeySpace[string]) *GetCommand {
	return &GetCommand{kp: kp}
}

// GetArgs represents the arguments for the GetCommand.
// It contains the key to retrieve the value for.
type GetArgs struct {
	Key string `json:"key" validate:"required"`
}

// Execute runs the GetCommand with the provided arguments.
func (c *GetCommand) Execute(args GetArgs) (*Output[string], errors.Error) {
	value, err := c.kp.Get(keys.Key(args.Key))
	if err != nil {
		return nil, NewCommandError("get-failed", "failed to get value", err).WithMetadata("args", args)
	}

	return Exit(value), nil
}

func (c *GetCommand) Name() string {
	return "get"
}

var _ Command[GetArgs, string] = (*GetCommand)(nil)
