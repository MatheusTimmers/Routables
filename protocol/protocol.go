package protocol

import (
	"fmt"
)

func FormatRoutingMessage(routeTable map[string]int) string {
	var message string = ""

	for ip, metric := range routeTable {
		message += fmt.Sprintf("!%s:%d", ip, metric)
	}
	return message
}
