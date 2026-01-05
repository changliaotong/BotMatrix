package policy

var CenterActionWhitelist = map[string]bool{
	"send_message":      true,
	"send_notification": true,
	"update_config":     true,
	"restart_plugin":    true,
	"stop_plugin":       true,
	"list_plugins":      true,
	"storage.get":       true,
	"storage.set":       true,
	"storage.delete":    true,
	"storage.exists":    true,
}
