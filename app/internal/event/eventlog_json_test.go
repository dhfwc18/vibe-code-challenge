package event_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vibe-code-challenge/twenty-fifty/internal/event"
)

// TestEventLog_MarshalJSON_RoundTrip verifies JSON encode/decode preserves entries.
func TestEventLog_MarshalJSON_RoundTrip(t *testing.T) {
	log := event.NewEventLog()
	log = event.AppendEventLog(log, event.EventEntry{DefID: "e1", Name: "First", Week: 1})
	log = event.AppendEventLog(log, event.EventEntry{DefID: "e2", Name: "Second", Week: 2})

	data, err := json.Marshal(log)
	assert.NoError(t, err)

	var restored event.EventLog
	assert.NoError(t, json.Unmarshal(data, &restored))

	entries := restored.Entries()
	assert.Len(t, entries, 2)
	assert.Equal(t, "First", entries[0].Name)
	assert.Equal(t, "Second", entries[1].Name)
	assert.Equal(t, 1, entries[0].Week)
}

// TestEventLog_MarshalJSON_Empty verifies an empty log marshals to "[]".
func TestEventLog_MarshalJSON_Empty(t *testing.T) {
	log := event.NewEventLog()
	data, err := json.Marshal(log)
	assert.NoError(t, err)
	assert.Equal(t, "null", string(data))
}

// TestEventLog_UnmarshalJSON_Empty verifies unmarshalling "[]" produces no entries.
func TestEventLog_UnmarshalJSON_Empty(t *testing.T) {
	var log event.EventLog
	assert.NoError(t, json.Unmarshal([]byte("[]"), &log))
	assert.Nil(t, log.Entries())
}
