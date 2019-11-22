package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormattedEventAction(t *testing.T) {
	assert := assert.New(t)
	event := types.FixtureEvent("entity1", "check1")

	action := formattedEventAction(event)
	assert.Equal("RESOLVED", action)

	event.Check.Status = 1
	action = formattedEventAction(event)
	assert.Equal("ALERT", action)
}

func TestChomp(t *testing.T) {
	assert := assert.New(t)

	trimNewline := chomp("hello\n")
	assert.Equal("hello", trimNewline)

	trimCarriageReturn := chomp("hello\r")
	assert.Equal("hello", trimCarriageReturn)

	trimBoth := chomp("hello\r\n")
	assert.Equal("hello", trimBoth)

	trimLots := chomp("hello\r\n\r\n\r\n")
	assert.Equal("hello", trimLots)
}

func TestEventKey(t *testing.T) {
	assert := assert.New(t)
	event := types.FixtureEvent("entity1", "check1")
	eventKey := eventKey(event)
	assert.Equal("entity1/check1", eventKey)
}

func TestEventSummary(t *testing.T) {
	assert := assert.New(t)
	event := types.FixtureEvent("entity1", "check1")
	event.Check.Output = "disk is full"

	eventKey := eventSummary(event, 100)
	assert.Equal("entity1/check1:disk is full", eventKey)

	eventKey = eventSummary(event, 5)
	assert.Equal("entity1/check1:disk ...", eventKey)
}

func TestFormattedMessage(t *testing.T) {
	assert := assert.New(t)
	event := types.FixtureEvent("entity1", "check1")
	event.Check.Output = "disk is full"
	event.Check.Status = 1
	formattedMsg := formattedMessage(event)
	assert.Equal("ALERT - entity1/check1:disk is full", formattedMsg)
}

func TestStateToEmojifier(t *testing.T) {
	assert := assert.New(t)
	event := types.FixtureEvent("entity1", "check1")

	event.Check.Status = 0
	color, state, icon, _ := stateToEmojifier(event)
	assert.Equal("success", color)
	assert.Equal("Resolved", state)
	assert.Equal("✅", icon)

	event.Check.Status = 1
	color, state, icon, _ = stateToEmojifier(event)
	assert.Equal("warning", color)
	assert.Equal("Warning", state)
	assert.Equal("️⚠️", icon)

	event.Check.Status = 2
	color, state, icon, _ = stateToEmojifier(event)
	assert.Equal("danger", color)
	assert.Equal("Critical", state)
	assert.Equal("🚨", icon)

	event.Check.Status = 33
	color, state, icon, _ = stateToEmojifier(event)
	assert.Equal("unknown", color)
	assert.Equal("Unknown", state)
	assert.Equal("⁉️", icon)
}

func TestMain(t *testing.T) {
	assert := assert.New(t)
	file, _ := ioutil.TempFile(os.TempDir(), "sensu-handler-webex-")
	defer func() {
		_ = os.Remove(file.Name())
	}()

	event := types.FixtureEvent("entity1", "check1")
	eventJSON, _ := json.Marshal(event)
	_, err := file.WriteString(string(eventJSON))
	require.NoError(t, err)
	require.NoError(t, file.Sync())
	_, err = file.Seek(0, 0)
	require.NoError(t, err)
	stdin = file

	requestReceived := true

	oldArgs := os.Args
	os.Args = []string{"webex"} ///, "-w", apiStub.URL}
	defer func() { os.Args = oldArgs }()

	main()

	assert.True(requestReceived)
}
