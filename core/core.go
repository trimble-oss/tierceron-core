package core

const (
	PLUGIN_EVENT_START = iota
	PLUGIN_EVENT_STOP
	PLUGIN_EVENT_STATUS
)

const PLUGIN_EVENT_CHANNELS_MAP_KEY = "PluginEventChannelsMap"
const PLUGIN_CHANNEL_EVENT_IN = "PluginChannelEventIn"
const PLUGIN_CHANNEL_EVENT_OUT = "PluginChannelEventOut"
const DATA_FLOW_STAT_CHANNEL = "DataFlowStatisticsChannel"
const ERROR_CHANNEL = "ErrorChannel"

const RFC_ISO_8601 = "2006-01-02 15:04:05 -0700 MST"
