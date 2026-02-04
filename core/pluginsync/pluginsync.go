package pluginsync

import "github.com/glycerine/bchan"

// pluginReadyChannels coordinates plugin startup synchronization using broadcast channels
var pluginReadyChannels = make(map[string]*bchan.Bchan)

// CreatePluginReadyChannel creates a broadcast channel for a plugin to signal when it's ready.
// Multiple goroutines can wait on the same channel and all will be notified.
func CreatePluginReadyChannel(pluginName string) *bchan.Bchan {
	ch := bchan.New(1) // Buffer size 1 for broadcast
	pluginReadyChannels[pluginName] = ch
	return ch
}

// SignalPluginReady signals that a plugin has completed initialization and is ready.
// This broadcasts to all waiting goroutines.
func SignalPluginReady(pluginName string) {
	if ch, ok := pluginReadyChannels[pluginName]; ok {
		ch.Bcast(true)
	}
}

// WaitForPluginReady blocks until the specified plugin signals it's ready.
// If the plugin channel doesn't exist, this returns immediately.
// Multiple goroutines can wait on the same plugin simultaneously.
func WaitForPluginReady(pluginName string) {
	if ch, ok := pluginReadyChannels[pluginName]; ok {
		<-ch.Ch
		ch.BcastAck()
	}
}
