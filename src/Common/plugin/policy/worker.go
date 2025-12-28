package policy

var WorkerActionWhitelist = map[string]bool{
	"send_message":      true,
	"send_notification": true,
	"record_data":       true,
	"query_data":        true,
}
