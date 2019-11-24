package main

import (
	"testing"
	"time"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

var (
	eventWithStatus = &corev2.Event{
		Check: &corev2.Check{
			Output: "This is a string w/ new line\n",
			Status: 10,
		},
	}
)

func TestTrimSuffixCheckOutput(t *testing.T) {
	assert := assert.New(t)
	output := trimSuffixCheckOutput(eventWithStatus)
	assert.Equal("This is a string w/ new line", output)
}

func TestParseTime(t *testing.T) {
	assert := assert.New(t)
	timeStamp, _ := time.Parse("Jan 2, 2006 at 3:04pm (MST)", "Dec 29, 2014 at 7:54pm (SGT)")
	assert.Equal("2014-12-29 19:54:00 +0000 SGT", timeStamp.String())
}

// func TestCheck(t *testing.T) {
// 	newCheck := corev2.FixtureCheck("check")
// 	newCheck.History = []corev2.CheckHistory{
// 		Status: 1,
// 	}

// 	assert := assert.New(t)
// 	seed := time.Now().UnixNano()

// 	popr := math_rand.New(math_rand.NewSource(seed))
// 	p := corev2.NewPopulatedCheckHistory(popr, false)

// 	// err := validateEvent(&corev2.Event{
// 	// 	Timestamp: 11231231,

// 	// 	Check: &corev2.Check{
// 	// 		Output:  "This is a string w/ new line\n",
// 	// 		History: [p],
// 	// 	},
// 	// })
// 	assert.Nil(err)
// }

func TestValidateEvent(t *testing.T) {

	assert := assert.New(t)

	err := validateEvent(&corev2.Event{
		Timestamp: 11231231,
		Entity: &corev2.Entity{
			EntityClass: "agent",
			ObjectMeta: corev2.ObjectMeta{
				Name:      "fp",
				Namespace: "default",
			},
		},
		Check: &corev2.Check{
			ObjectMeta: corev2.ObjectMeta{
				Name: "check-name",
			},
			Interval: 20,
		},
	})
	assert.Nil(err)

	assert.EqualError(validateEvent(&corev2.Event{}), "timestamp is missing or must be greater than zero")
	assert.EqualError(validateEvent(&corev2.Event{
		Timestamp: 11231231,
	}), "entity is missing from event")

	assert.EqualError(validateEvent(&corev2.Event{
		Timestamp: 11231231,
		Entity:    &corev2.Entity{},
	}), "check is missing from event")

	assert.EqualError(validateEvent(&corev2.Event{
		Timestamp: 11231231,
		Entity: &corev2.Entity{
			ObjectMeta: corev2.ObjectMeta{
				Name: "fp",
			},
		},
		Check: &corev2.Check{},
	}), "entity class must not be empty")

	assert.EqualError(validateEvent(&corev2.Event{
		Timestamp: 11231231,
		Entity: &corev2.Entity{
			EntityClass: "agent",
			ObjectMeta: corev2.ObjectMeta{
				Name:      "fp",
				Namespace: "default",
			},
		},
		Check: &corev2.Check{
			ObjectMeta: corev2.ObjectMeta{
				Name: "check-name",
			},
		},
	}), "check interval must be greater than or equal to 1")

}

func TestCheckArgs(t *testing.T) {
	assert := assert.New(t)
	err := checkArgs(&corev2.Event{
		Timestamp: 11231231,
		Entity: &corev2.Entity{
			EntityClass: "agent",
			ObjectMeta: corev2.ObjectMeta{
				Name:      "fp",
				Namespace: "default",
			},
		},
		Check: &corev2.Check{
			Output:   "This is a string w/ new line\n",
			Status:   10,
			Interval: 20,
			ObjectMeta: corev2.ObjectMeta{
				Name: "check-name",
			},
		},
	})
	assert.Nil(err)

	assert.EqualError(checkArgs(&corev2.Event{}), "event does not contain check")

}
func TestStringMinifier(t *testing.T) {
	assert := assert.New(t)

	input := `<blockquote class='blue'> foo bar <br/>
      <b>Check Name:</b> foocheck
      <b>Entity:</b> entity-nmame      <br/>
      <b>Check output:</b>  this is output <br/>
      <b>History:</b>  <br/>
	</blockquote>`

	output := stringMinifier(input)
	assert.Equal(output, "<blockquote class='blue'> foo bar <br/> <b>Check Name:</b> foocheck <b>Entity:</b> entity-nmame <br/> <b>Check output:</b> this is output <br/> <b>History:</b> <br/> </blockquote>")

}
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
