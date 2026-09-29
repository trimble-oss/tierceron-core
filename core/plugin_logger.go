package core

import (
	"io"
	"log"
	"strings"
	"sync"
)

var activeLogPlugins sync.Map

type pluginLogWriter struct {
	pluginName string
	writer     io.Writer
}

func (writer *pluginLogWriter) Write(message []byte) (int, error) {
	if active, ok := activeLogPlugins.Load(writer.pluginName); ok && !active.(bool) {
		return len(message), nil
	}
	return writer.writer.Write(message)
}

func ActivatePluginLogging(pluginName string) {
	if pluginName != "" {
		activeLogPlugins.LoadOrStore(pluginName, true)
	}
}

func SetPluginLoggingActive(pluginName string, active bool) {
	if pluginName != "" {
		activeLogPlugins.Store(pluginName, active)
	}
}

func IsPluginLoggingActive(pluginName string) bool {
	active, ok := activeLogPlugins.Load(pluginName)
	return !ok || active.(bool)
}

func NewPluginLogger(pluginName string, logger *log.Logger) *log.Logger {
	ActivatePluginLogging(pluginName)
	if logger == nil || pluginName == "" {
		return logger
	}
	prefix := logger.Prefix()
	if !strings.Contains(prefix, "["+pluginName+"]") {
		prefix += "[" + pluginName + "] "
	}
	return log.New(&pluginLogWriter{pluginName: pluginName, writer: logger.Writer()}, prefix, logger.Flags())
}
