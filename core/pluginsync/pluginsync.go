package pluginsync

import (
	"sync"

	"github.com/glycerine/bchan"
)

// pluginReadyChannels coordinates plugin startup synchronization using broadcast channels
var pluginReadyChannels = make(map[string]*bchan.Bchan)
var pluginReadyMu sync.RWMutex

// CreatePluginReadyChannel creates a broadcast channel for a plugin to signal when it's ready.
// Multiple goroutines can wait on the same channel and all will be notified.
func CreatePluginReadyChannel(pluginName string) *bchan.Bchan {
	ch := bchan.New(1) // Buffer size 1 for broadcast
	pluginReadyMu.Lock()
	pluginReadyChannels[pluginName] = ch
	pluginReadyMu.Unlock()
	return ch
}

// IsPluginDeployed reports whether the kernel registered a local plugin deployment.
func IsPluginDeployed(pluginName string) bool {
	return pluginReadyChannel(pluginName) != nil
}

func pluginReadyChannel(pluginName string) *bchan.Bchan {
	pluginReadyMu.RLock()
	defer pluginReadyMu.RUnlock()
	return pluginReadyChannels[pluginName]
}

// SignalPluginReady signals that a plugin has completed initialization and is ready.
// This broadcasts to all waiting goroutines.
func SignalPluginReady(pluginName string) {
	if ch := pluginReadyChannel(pluginName); ch != nil {
		ch.Bcast(true)
	}
}

// WaitForPluginReady blocks until the specified plugin signals it's ready.
// If the plugin channel doesn't exist, this returns immediately.
// Multiple goroutines can wait on the same plugin simultaneously.
func WaitForPluginReady(pluginName string) {
	if ch := pluginReadyChannel(pluginName); ch != nil {
		<-ch.Ch
		ch.BcastAck()
	}
}
