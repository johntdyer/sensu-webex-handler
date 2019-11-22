package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"
	"unicode"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-plugins-go-library/sensu"

	"github.com/alecthomas/template"
	"github.com/sensu/sensu-go/types"

	"github.com/go-resty/resty/v2"
)

const (
	endpointURL = "endpoint-url"
	apiKey      = "api-key"
	roomID      = "room-id"
	apiURL      = "api-url"

	staticImageBucket = "https://webex-teams-static-image-store.s3.us-east-2.amazonaws.com"
)

// HandlerConfig represents plugin configuration settings.
type HandlerConfig struct {
	sensu.PluginConfig
	Timeout           int
	WebexTeamsRoomID  string
	WebexTeamsAuthKey string
	WebexTeamsAPI     string
}

var (
	config = HandlerConfig{
		PluginConfig: sensu.PluginConfig{
			Name:     "sensu-webex-teams-handler",
			Short:    "The Sensu Go Webex Teams handler for event forwarding",
			Timeout:  10,
			Keyspace: "sensu.io/plugins/webex-teams/config",
		},
	}
	teamnsConfigOptions = []*sensu.PluginConfigOption{
		{
			Path:      roomID,
			Env:       "WEBEX_TEAMS_ROOM_ID",
			Argument:  roomID,
			Shorthand: "r",
			Default:   "",
			Usage:     "Room or email to send alert to",
			Value:     &config.WebexTeamsRoomID,
		},
		{
			Path:      apiKey,
			Env:       "WEBEX_TEAMS_API_KEY",
			Argument:  apiKey,
			Shorthand: "k",
			Default:   "",
			Usage:     "API key for authenticated access",
			Value:     &config.WebexTeamsAuthKey,
		},
		{
			Path:      apiURL,
			Env:       "WEBEX_TEAMS_API_URL",
			Argument:  apiURL,
			Shorthand: "a",
			Default:   "api.ciscospark.com",
			Usage:     "Webex Teams API",
			Value:     &config.WebexTeamsAPI,
		},
	}
)

func main() {
	goHandler := sensu.NewGoHandler(&config.PluginConfig, teamnsConfigOptions, checkArgs, sendMessage)
	goHandler.Execute()
}

func checkArgs(event *corev2.Event) error {
	if !event.HasCheck() {
		return fmt.Errorf("event does not contain check")
	}
	return nil
}

func formattedEventAction(event *corev2.Event) string {
	switch event.Check.Status {
	case 0:
		return "RESOLVED"
	default:
		return "ALERT"
	}
}

func chomp(s string) string {
	return strings.Trim(strings.Trim(strings.Trim(s, "\n"), "\r"), "\r\n")
}

func eventKey(event *corev2.Event) string {
	return fmt.Sprintf("%s/%s", event.Entity.Name, event.Check.Name)
}

func eventSummary(event *corev2.Event, maxLength int) string {
	output := chomp(event.Check.Output)
	if len(event.Check.Output) > maxLength {
		output = output[0:maxLength] + "..."
	}
	return fmt.Sprintf("%s:%s", eventKey(event), output)
}

func formattedMessage(event *corev2.Event) string {
	return fmt.Sprintf("%s - %s", formattedEventAction(event), eventSummary(event, 100))
}

// stringMinifier remove whitespace before sending message to teams
func stringMinifier(in string) (out string) {
	white := false
	for _, c := range in {
		if unicode.IsSpace(c) {
			if !white {
				out = out + " "
			}
			white = true
		} else {
			out = out + string(c)
			white = false
		}
	}
	return
}

func stateToEmojifier(event *corev2.Event) (string, string, string, uint32) {
	switch event.Check.Status {
	case 0:
		return "success", "Resolved", "✅", event.Check.Status
	case 2:
		return "danger", "Critical", "🚨", event.Check.Status
	case 1:
		return "warning", "Warning", "️⚠️", event.Check.Status
	default:
		return "unknown", "Unknown", "⁉️", event.Check.Status
	}
}

//Sometimes the check output ends ina string and this mucks w/ the template
func trimSuffixCheckOutput(event *corev2.Event) string {

	text := strings.TrimSuffix(event.Check.Output, "\n")
	return text
}

// Take event list and ensure it is sorted based on timestamp, then return list
func getSortedHistory(event *corev2.Event) []string {

	history := []types.CheckHistory{}
	for _, record := range event.Check.History {
		latest := types.CheckHistory{
			Executed: record.Executed,
			Status:   record.Status,
		}
		history = append(history, latest)
	}

	// Sort
	sort.Sort(types.ByExecuted(history))

	statusHistory := []string{}
	for _, entry := range history {
		statusHistory = append(statusHistory, fmt.Sprint(entry.Status))
	}

	return statusHistory
}

func parseTime(input time.Time) string {
	return input.Format("Monday 01/02/2006 - 15:04:05 MST")
}

func getTemplateNew(event *corev2.Event) string {

	templateWithoutNewlines := stringMinifier(inccidentTemplate)

	card := template.Must(template.New("inccident").Funcs(template.FuncMap{
		"parseTime": parseTime,
	}).Parse(templateWithoutNewlines))

	var tpl bytes.Buffer

	eventColor, eventState, emoji, _ := stateToEmojifier(event)

	var messageTarget string

	if strings.Contains(config.WebexTeamsRoomID, "@") {
		messageTarget = fmt.Sprintf("\"toPersonEmail\":\"%s\",", config.WebexTeamsRoomID)
	} else {
		messageTarget = fmt.Sprintf("\"roomId\":\"%s\",", config.WebexTeamsRoomID)
	}

	localStruct := struct {
		CheckOutput          string
		CheckExecutionTime   time.Time
		MessageColor         string
		MessageStatus        string
		MessageTarget        string
		EntityName           string
		CheckName            string
		Emoji                string
		History              []string
		FormattedEventAction string
		FormattedMessage     string
		BucketName           string
	}{

		strings.TrimSuffix(event.Check.GetOutput(), "\n"),
		time.Unix(event.Check.GetExecuted(), 0),
		eventColor,
		eventState,
		messageTarget,
		event.Entity.Name,
		event.Check.GetObjectMeta().Name,
		emoji,
		getSortedHistory(event),
		formattedEventAction(event),
		formattedMessage(event),
		staticImageBucket,
	}

	err := card.Execute(&tpl, localStruct)
	if err != nil {
		panic(err)
	}

	return tpl.String()

}

func sendMessage(event *corev2.Event) error {
	client := resty.New()

	validateEvent(event)

	template := getTemplateNew(event)

	client.
		SetRetryCount(2).
		SetRetryWaitTime(2 * time.Second).
		SetRetryMaxWaitTime(9 * time.Second).
		SetRetryAfter(func(client *resty.Client, resp *resty.Response) (time.Duration, error) {
			return 0, errors.New("quota exceeded")
		})

	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(template).
		SetAuthToken(config.WebexTeamsAuthKey).
		Post("https://" + config.WebexTeamsAPI + "/v1/messages")

	if err != nil {
		log.Fatal(err)
	}

	if resp.StatusCode() == 200 {
		return nil
	}

	fmt.Println("Template: ")
	fmt.Println(template)
	fmt.Println("")
	fmt.Println("Response Info:")

	fmt.Println("Status Code:", resp.StatusCode())
	fmt.Println("Status     :", resp.Status())
	fmt.Println("Time       :", resp.Time())
	fmt.Println("Received At:", resp.ReceivedAt())
	fmt.Println("Body       :\n", resp)
	fmt.Println()
	return nil
}

func validateEvent(event *corev2.Event) error {
	if event.Timestamp <= 0 {
		return errors.New("timestamp is missing or must be greater than zero")
	}

	if event.Entity == nil {
		return errors.New("entity is missing from event")
	}

	if !event.HasCheck() {
		return errors.New("check is missing from event")
	}

	if err := event.Entity.Validate(); err != nil {
		return err
	}

	if err := event.Check.Validate(); err != nil {
		return errors.New(err.Error())
	}

	return nil
}
