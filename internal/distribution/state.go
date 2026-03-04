package distribution

import (
	"encoding/json"
	"io"

	"github.com/William-Fernandes252/clavis/internal/commands"
	"github.com/William-Fernandes252/clavis/internal/keys"
	"github.com/hashicorp/raft"
)

// FSM is a Raft finite state machine that applies commands to the key-value store.
type FSM struct {
	ks keys.KeySpace[*keys.Entry]
}

// Apply applies a command to the key-value store.
func (f *FSM) Apply(log *raft.Log) interface{} {
	var cmd commands.Command[any]
	if err := json.Unmarshal(log.Data, &cmd); err != nil {
		panic("failed to unmarshal command: " + err.Error())
	}

	// For now, we are ignoring the output of the command.
	// In a real-world scenario, you might want to handle this differently.
	_, err := cmd.Execute()
	return err
}

// Snapshot returns a snapshot of the current state of the FSM.
func (f *FSM) Snapshot() (raft.FSMSnapshot, error) {
	// This is a simplified implementation. In a real-world scenario, you would
	// need to create a snapshot of your key-value store.
	return &snapshot{}, nil
}

// Restore restores the FSM to a previous state.
func (f *FSM) Restore(rc io.ReadCloser) error {
	// This is a simplified implementation. In a real-world scenario, you would
	// need to restore your key-value store from a snapshot.
	return nil
}

type snapshot struct{}

func (s *snapshot) Persist(sink raft.SnapshotSink) error {
	return sink.Close()
}

func (s *snapshot) Release() {}
