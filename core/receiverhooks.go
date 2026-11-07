package core

import (
	"errors"
	"fmt"
	"log"
	"time"

	cmap "github.com/orcaman/concurrent-map/v2"
)

type ChatHookFunc func(*ChatMsg) bool

var trcdbQueryLoggingActive = cmap.New[bool]()

func SetTrcdbQueryLoggingActive(sourcePlugin string, active bool) {
	trcdbQueryLoggingActive.Set(sourcePlugin, active)
}

func IsTrcdbQueryLoggingActive(sourcePlugin string) bool {
	active, _ := trcdbQueryLoggingActive.Get(sourcePlugin)
	return active
}

func RegisterChatMsgHook(chatMsgHookCtx **cmap.ConcurrentMap[string, ChatHookFunc], id string, hook ChatHookFunc) (string, error) {
	if chatMsgHookCtx == nil {
		return "", errors.New("chatMsgHookCtx pointer is nil")
	}
	if *chatMsgHookCtx == nil {
		cmHooks := cmap.New[ChatHookFunc]()
		*chatMsgHookCtx = &cmHooks
	}
	if hook == nil {
		return "", errors.New("hook cannot be nil")
	}

	for {
		if (*chatMsgHookCtx).SetIfAbsent(id, hook) {
			break
		} else {
			// Probably never will trigger...  But just in case..
			id = fmt.Sprintf("%s-%d", id, time.Now().UnixNano())
		}
	}
	return id, nil
}

func CallChatMsgHooks(chatMsgHooks *cmap.ConcurrentMap[string, ChatHookFunc], msg *ChatMsg) {
	handled := false
	for _, hook := range chatMsgHooks.Items() {
		handled = handled || hook(msg)
	}
	if handled {
		chatMsgHooks.Remove(*msg.ChatId)
	}
}

func CallSelectedChatMsgHook(chatMsgHooks *cmap.ConcurrentMap[string, ChatHookFunc], msg *ChatMsg) {
	handled := false
	if hook, ok := chatMsgHooks.Get(*msg.ChatId); ok {
		handled = handled || hook(msg)
	}
	if handled {
		chatMsgHooks.Remove(*msg.ChatId)
	}
}

func CallChatQueryChan(chatMsgHookCtx **cmap.ConcurrentMap[string, ChatHookFunc],
	sourcePlugin string,
	trcdbExchange *TrcdbExchange,
	chatSenderChan *chan *ChatMsg, loggers ...*log.Logger,
) *ChatMsg {
	var logger *log.Logger
	if len(loggers) > 0 {
		logger = loggers[0]
	}
	id := fmt.Sprintf("%s-%d", sourcePlugin, time.Now().UnixNano())
	chatTrcdbQueryMsg := ChatMsg{
		ChatId: &id,
	}
	name := sourcePlugin
	chatTrcdbQueryMsg.Name = &name
	chatTrcdbQueryMsg.Query = &[]string{"trcdb"}
	chatTrcdbQueryMsg.TrcdbExchange = trcdbExchange
	responseChan := make(chan *ChatMsg, 1)
	newId, _ := RegisterChatMsgHook(chatMsgHookCtx, id, func(msg *ChatMsg) bool {
		if msg.ChatId != nil && *msg.ChatId == id {
			if msg.TrcdbExchange != nil {
				go func() {
					responseChan <- msg
				}()
				return true
			}
		}
		return false
	})
	if newId != id {
		chatTrcdbQueryMsg.ChatId = &newId
	}

	go func() {
		if logger != nil && IsTrcdbQueryLoggingActive(sourcePlugin) {
			query := ""
			if trcdbExchange != nil {
				query = trcdbExchange.Query
			}
			logger.Printf("Sending Trcdb query: source=%s id=%s query=%q", sourcePlugin, newId, query)
		}
		*chatSenderChan <- &chatTrcdbQueryMsg
	}()

	chatResponseMsg := <-responseChan
	if logger != nil && IsTrcdbQueryLoggingActive(sourcePlugin) {
		logger.Printf("Received Trcdb query response: source=%s id=%s", sourcePlugin, newId)
	}
	return chatResponseMsg
}
