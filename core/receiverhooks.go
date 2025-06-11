package core

import (
	cmap "github.com/orcaman/concurrent-map/v2"
)

type ChatHookFunc func(*ChatMsg) bool

func RegisterChatMsgHook(chatMsgHookCtx *cmap.ConcurrentMap[string, ChatHookFunc], id string, hook ChatHookFunc) {
	if chatMsgHookCtx == nil {
		cmHooks := cmap.New[ChatHookFunc]()
		chatMsgHookCtx = &cmHooks
	}
	if hook == nil {
		return
	}

	chatMsgHookCtx.Set(id, hook)
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
