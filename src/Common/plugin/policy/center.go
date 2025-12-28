package policy

var CenterActionWhitelist = map[string]bool{
	"send_message":      true,
	"send_notification": true,
	"update_config":     true,
	"restart_plugin":    true,
	"stop_plugin":       true,
	"list_plugins":      true,
}
