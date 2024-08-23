
## Let's get started
When implementing plugins several things must be adhered to:
* Correct build options.
* Go module dependencies must match versions with those in github.com/trimble-oss/tierceron

## Intialize a function
Only 2 functions needs to be implemented in the plugin:

## Provide paths in vault for certificates
func GetConfigPaths() []string

## Provide secrets, certificates, and an event channel map 
func Init(properties *map[string]interface{})

# Keys include:
PluginEventChannelsMap: map[string]interface{}
Keys provided by GetConfigPaths
Built in project/service key containing project secrets

# Keys in PluginEventChannelsMap include:
PluginChannelEventIn  - contains channel for receving incoming events
PluginChannelEventOut - channel for emitting event results



