package core

import (
	"errors"
	"fmt"
	"time"

	cmap "github.com/orcaman/concurrent-map/v2"
)

type ChatHookFunc func(*ChatMsg) bool

func RegisterChatMsgHook(chatMsgHookCtx *cmap.ConcurrentMap[string, ChatHookFunc], id string, hook ChatHookFunc) (string, error) {
	if chatMsgHookCtx == nil {
		cmHooks := cmap.New[ChatHookFunc]()
		chatMsgHookCtx = &cmHooks
	}
	if hook == nil {
		return "", errors.New("hook cannot be nil")
	}

	for {
		if chatMsgHookCtx.SetIfAbsent(id, hook) {
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
