package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/alecthomas/template"
	webexteams "github.com/jbogarin/go-cisco-webex-teams/sdk"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
	resty "gopkg.in/resty.v1"
)

var (
	roomID string

	token string

	timeout int
	stdin   *os.File
)

func main() {
	rootCmd := configureRootCommand()
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err.Error())
	}
}

func configureRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sensu-webex-handler",
		Short: "The Sensu Go Webex Teams handler for notifying a channel",
		RunE:  run,
	}

	cmd.Flags().StringVarP(&roomID,
		"room-id",
		"r",
		"",
		"The space to post messages to, can also be a users email if you want to send directly to a person vs a space")

	cmd.Flags().StringVarP(&token,
		"token",
		"",
		"",
		"the api token to use")

	cmd.Flags().IntVarP(&timeout,
		"timeout",
		"t",
		10,
		"The amount of seconds to wait before terminating the handler")

	return cmd
}

func run(cmd *cobra.Command, args []string) error {
	if len(args) != 0 {
		_ = cmd.Help()
		return errors.New("invalid argument(s) received")
	}
	if stdin == nil {
		stdin = os.Stdin
	}

	eventJSON, err := ioutil.ReadAll(stdin)
	if err != nil {
		return fmt.Errorf("failed to read stdin: %s", err.Error())
	}

	event := &types.Event{}
	err = json.Unmarshal(eventJSON, event)
	if err != nil {
		return fmt.Errorf("failed to unmarshal stdin data: %s", eventJSON)
	}

	if err = validateEvent(event); err != nil {
		return errors.New(err.Error())
	}

	if err = sendMessage(event); err != nil {
		return errors.New(err.Error())
	}

	return nil
}

func formattedEventAction(event *types.Event) string {
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

func eventKey(event *types.Event) string {
	return fmt.Sprintf("%s/%s", event.Entity.Name, event.Check.Name)
}

func eventSummary(event *types.Event, maxLength int) string {
	output := chomp(event.Check.Output)
	if len(event.Check.Output) > maxLength {
		output = output[0:maxLength] + "..."
	}
	return fmt.Sprintf("%s:%s", eventKey(event), output)
}

func formattedMessage(event *types.Event) string {
	return fmt.Sprintf("%s - %s", formattedEventAction(event), eventSummary(event, 100))
}

func messageColor(event *types.Event) string {
	switch event.Check.Status {
	case 0:
		return "success"
	case 2:
		return "danger"
	default:
		return "warning"
	}
}

func messageStatus(event *types.Event) string {
	switch event.Check.Status {
	case 0:
		return "Resolved"
	case 2:
		return "Critical"
	default:
		return "Warning"
	}
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

func stateToEmojifier(event *types.Event) (string, string, uint32) {
	switch event.Check.Status {
	case 0:
		return "success", "✅", event.Check.Status
	case 2:
		return "danger", "🚨", event.Check.Status
	case 1:
		return "warning", "️⚠️", event.Check.Status
	default:
		return "unknown", "⁉️", event.Check.Status
	}
}

// Take event list and ensure it is sorted based on timestamp, then return list
func getSortedHistory(event *types.Event) []string {

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

func getTemplateNew(condition string, event *types.Event) string {

	templateWithoutNewlines := stringMinifier(inccidentTemplate)

	t := template.Must(template.New("inccident").Funcs(template.FuncMap{
		"parseTime": parseTime,
	}).Parse(templateWithoutNewlines))

	var tpl bytes.Buffer

	eventColor, emoji, _ := stateToEmojifier(event)

	// fmt.Println(getSortedHistory(event))
	// history := []types.CheckHistory{}

	// fmt.Printf("%# v", pretty.Formatter(event))

	localStruct := struct {
		CheckOutput          string
		CheckExecutionTime   time.Time
		MessageColor         string
		MessageStatus        string
		EntityName           string
		CheckName            string
		Emoji                string
		History              []string
		FormattedEventAction string
		FormattedMessage     string
	}{
		event.Check.GetOutput(),
		time.Unix(event.Check.GetExecuted(), 0),
		eventColor, //messageColor(event),
		messageStatus(event),
		event.Entity.Name,
		event.Entity.GetObjectMeta().Name,
		emoji,
		getSortedHistory(event),
		formattedEventAction(event),
		formattedMessage(event),
	}

	err := t.Execute(&tpl, localStruct)
	if err != nil {
		panic(err)
	}

	return tpl.String()

}

// Define a template.
const inccidentTemplate = `
<blockquote class='{{.MessageColor}}'> {{.Emoji}} {{.MessageStatus}} <br/>
<b>Check Name:</b> {{.CheckName}} 
{{if (ne .MessageStatus "Resolved") }}
&nbsp;&nbsp;&nbsp;&nbsp; <b>Execution Time:</b> {{.CheckExecutionTime}} <br/>
{{end}}
<b>Entity:</b> {{.EntityName}}      <br/>
{{if (ne .MessageStatus "Resolved") }}
<b>Check output:</b> {{.CheckOutput}} <br/>
<b>History:</b> {{.History}} <br/>
{{end}}
</blockquote>
`

func sendMessage(event *types.Event) error {

	client := resty.New()

	client.SetAuthToken(token)
	Client := webexteams.NewClient(client)

	template := getTemplateNew(messageStatus(event), event)

	markDownMessage := &webexteams.MessageCreateRequest{
		Markdown: template,
	}

	// If room id contains an @ we assume email address and use that
	if strings.Contains(roomID, "@") {
		markDownMessage.ToPersonEmail = roomID

	} else {
		markDownMessage.RoomID = roomID
	}

	newMarkDownMessage, _, err := Client.Messages.CreateMessage(markDownMessage)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("POST:", newMarkDownMessage.ID, newMarkDownMessage.Markdown, newMarkDownMessage.Created)

	return nil
}

func validateEvent(event *types.Event) error {
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
