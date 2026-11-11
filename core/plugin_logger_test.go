package core

import (
	"bytes"
	"log"
	"strings"
	"testing"

	cmap "github.com/orcaman/concurrent-map/v2"
)

func TestPluginLoggerScopesAndControlsOutput(t *testing.T) {
	const pluginName = "plugin-logger-test"

	var output bytes.Buffer
	logger := NewPluginLogger(pluginName, log.New(&output, "[trcshk]", 0))
	logger.Println("enabled")
	if !strings.Contains(output.String(), "[trcshk][plugin-logger-test] enabled") {
		t.Fatalf("expected scoped output, got %q", output.String())
	}

	SetPluginLoggingActive(pluginName, false)
	output.Reset()
	logger.Println("disabled")
	if output.Len() != 0 {
		t.Fatalf("expected disabled plugin output to be suppressed, got %q", output.String())
	}
}

func TestPluginLoggingDirective(t *testing.T) {
	const pluginName = "plugin-directive-test"
	ActivatePluginLogging(pluginName)
	defer SetPluginLoggingActive(pluginName, true)

	chatSender := make(chan *ChatMsg, 1)
	configContext := &ConfigContext{PluginName: pluginName, ChatSenderChan: &chatSender}
	target := pluginName
	source := "trcshtalk"
	directive := "log stop"
	routingID := "directive-test"
	query := []string{source}
	if !handlePluginLoggingDirective(configContext, &ChatMsg{Name: &target, Query: &query, ChatId: &directive, RoutingId: &routingID}) {
		t.Fatal("expected logging directive to be handled")
	}
	if IsPluginLoggingActive(pluginName) {
		t.Fatal("expected logging directive to disable the plugin")
	}
	response := <-chatSender
	if response.Response == nil || *response.Response != "logging disabled for plugin-directive-test" {
		t.Fatalf("unexpected directive response: %#v", response)
	}
	if response.Name == nil || *response.Name != pluginName || response.Query == nil || len(*response.Query) != 1 || (*response.Query)[0] != source {
		t.Fatalf("unexpected response routing: %#v", response)
	}
}

func TestPluginLoggingDirectiveOnlyAffectsTargetPlugin(t *testing.T) {
	const pluginName = "trcshtalk-directive-test"
	ActivatePluginLogging(pluginName)
	defer SetPluginLoggingActive(pluginName, true)

	context := &ConfigContext{PluginName: pluginName}
	target := "rainier"
	source := "trcshtalk"
	query := []string{source}
	directive := "log stop"
	if handlePluginLoggingDirective(context, &ChatMsg{Name: &target, Query: &query, ChatId: &directive}) {
		t.Fatal("expected a non-target plugin to leave the event for its handler")
	}
	if !IsPluginLoggingActive(pluginName) {
		t.Fatal("expected TrcshTalk logging to remain enabled")
	}
}

func TestTrcshTalkCanDisableItsOwnLogging(t *testing.T) {
	const pluginName = "trcshtalk"
	ActivatePluginLogging(pluginName)
	defer SetPluginLoggingActive(pluginName, true)

	chatSender := make(chan *ChatMsg, 1)
	configContext := &ConfigContext{PluginName: pluginName, ChatSenderChan: &chatSender}
	target := pluginName
	source := "trcshtalk"
	query := []string{source}
	directive := "log stop"
	if !handlePluginLoggingDirective(configContext, &ChatMsg{Name: &target, Query: &query, ChatId: &directive}) {
		t.Fatal("expected TrcshTalk to handle its own logging directive")
	}
	if IsPluginLoggingActive(pluginName) {
		t.Fatal("expected TrcshTalk logging to be disabled")
	}

	acknowledgement := <-chatSender
	if handlePluginLoggingDirective(configContext, acknowledgement) {
		t.Fatal("expected the acknowledgement to pass through to the waiting talkback")
	}
	if IsPluginLoggingActive(pluginName) {
		t.Fatal("expected the acknowledgement not to re-enable TrcshTalk logging")
	}
}

func TestTrcdbQueryLoggingBySource(test *testing.T) {
	const sourcePlugin = "query-logging-test"
	const otherSourcePlugin = "query-logging-other-test"
	defer SetTrcdbQueryLoggingActive(sourcePlugin, false)

	sendQuery := func(sourcePlugin string) string {
		var output bytes.Buffer
		chatSender := make(chan *ChatMsg, 1)
		chatHooks := cmap.New[ChatHookFunc]()
		hookCtx := &chatHooks
		exchange := &TrcdbExchange{Operation: "SELECT", Query: "SELECT 1"}
		response := &ChatMsg{TrcdbExchange: exchange}
		go func() {
			request := <-chatSender
			response.ChatId = request.ChatId
			CallSelectedChatMsgHook(&chatHooks, response)
		}()
		if received := CallChatQueryChan(&hookCtx, sourcePlugin, exchange, &chatSender, log.New(&output, "", 0)); received != response {
			test.Fatal("expected the query response to be delivered")
		}
		return output.String()
	}

	if output := sendQuery(sourcePlugin); output != "" {
		test.Fatalf("expected Trcdb query logging to default off, got %q", output)
	}
	SetTrcdbQueryLoggingActive(sourcePlugin, true)
	output := sendQuery(sourcePlugin)
	if strings.Count(output, "Trcdb query") != 2 || !strings.Contains(output, `query="SELECT 1"`) {
		test.Fatalf("expected a query/response pair including SQL text, got %q", output)
	}
	if output := sendQuery(otherSourcePlugin); output != "" {
		test.Fatalf("expected another source to remain disabled, got %q", output)
	}
	SetTrcdbQueryLoggingActive(sourcePlugin, false)
	if output := sendQuery(sourcePlugin); output != "" {
		test.Fatalf("expected disabled Trcdb query logging to be suppressed, got %q", output)
	}
}
