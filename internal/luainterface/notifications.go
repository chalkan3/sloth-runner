package luainterface

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/yuin/gopher-lua"
)

// NotificationsModule provides notification functionalities to Lua scripts
type NotificationsModule struct{}

// NewNotificationsModule creates a new NotificationsModule
func NewNotificationsModule() *NotificationsModule {
	return &NotificationsModule{}
}

// Loader returns the Lua loader for the notifications module
func (mod *NotificationsModule) Loader(L *lua.LState) int {
	// Create the main notifications table
	notificationsTable := L.NewTable()

	// Create and register the slack submodule
	slackModule := L.NewTable()
	L.SetFuncs(slackModule, map[string]lua.LGFunction{
		"send": mod.sendSlackNotification,
	})
	notificationsTable.RawSetString("slack", slackModule)

	// Create and register the ntfy submodule
	ntfyModule := L.NewTable()
	L.SetFuncs(ntfyModule, map[string]lua.LGFunction{
		"send": mod.sendNtfyNotification,
	})
	notificationsTable.RawSetString("ntfy", ntfyModule)

	// Return the main notifications table
	L.Push(notificationsTable)
	return 1
}

// sendSlackNotification sends a message to a Slack webhook.
// Lua usage: notifications.slack.send({webhook_url="...", message="...", pipeline="...", error_details="..."})
func (mod *NotificationsModule) sendSlackNotification(L *lua.LState) int {
	tbl := L.CheckTable(1)
	webhookURL := tbl.RawGetString("webhook_url").String()
	message := tbl.RawGetString("message").String()
	pipeline := tbl.RawGetString("pipeline").String()
	errorDetails := tbl.RawGetString("error_details").String()

	if webhookURL == "" {
		L.Push(lua.LBool(false))
		L.Push(lua.LString("Slack webhook_url is required"))
		return 2
	}

	// Create a more structured message for Slack
	attachmentText := fmt.Sprintf("Pipeline: *%s*", pipeline)
	if errorDetails != "" {
		attachmentText += fmt.Sprintf("\n```%s```", errorDetails)
	}

	payload := map[string]interface{}{
		"text": message,
		"attachments": []map[string]interface{}{
			{
				"color": "#f2c744", // Gemini yellow
				"text":  attachmentText,
			},
		},
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		L.Push(lua.LBool(false))
		L.Push(lua.LString(fmt.Sprintf("Failed to marshal Slack payload: %v", err)))
		return 2
	}

	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		L.Push(lua.LBool(false))
		L.Push(lua.LString(fmt.Sprintf("Failed to create Slack request: %v", err)))
		return 2
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		L.Push(lua.LBool(false))
		L.Push(lua.LString(fmt.Sprintf("Failed to send Slack notification: %v", err)))
		return 2
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		L.Push(lua.LBool(false))
		L.Push(lua.LString(fmt.Sprintf("Slack API returned non-200 status: %s", resp.Status)))
		return 2
	}

	L.Push(lua.LBool(true))
	return 1
}

// sendNtfyNotification sends a message to an ntfy topic.
// Lua usage: notifications.ntfy.send({server="...", topic="...", message="...", title="..."})
func (mod *NotificationsModule) sendNtfyNotification(L *lua.LState) int {
	tbl := L.CheckTable(1)
	server := tbl.RawGetString("server").String()
	topic := tbl.RawGetString("topic").String()
	message := tbl.RawGetString("message").String()
	title := tbl.RawGetString("title").String()
	priority := tbl.RawGetString("priority").String()

	if server == "" || topic == "" {
		L.Push(lua.LBool(false))
		L.Push(lua.LString("ntfy server and topic are required"))
		return 2
	}

	url := fmt.Sprintf("%s/%s", server, topic)

	req, err := http.NewRequest("POST", url, bytes.NewBufferString(message))
	if err != nil {
		L.Push(lua.LBool(false))
		L.Push(lua.LString(fmt.Sprintf("Failed to create ntfy request: %v", err)))
		return 2
	}

	req.Header.Set("Content-Type", "text/plain")
	if title != "" {
		req.Header.Set("Title", title)
	}
	if priority != "" {
		req.Header.Set("Priority", priority)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		L.Push(lua.LBool(false))
		L.Push(lua.LString(fmt.Sprintf("Failed to send ntfy notification: %v", err)))
		return 2
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		L.Push(lua.LBool(false))
		L.Push(lua.LString(fmt.Sprintf("ntfy server returned non-200 status: %s", resp.Status)))
		return 2
	}

	L.Push(lua.LBool(true))
	return 1
}
