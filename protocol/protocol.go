package protocol

import (
	"fmt"
	"routables/router"
	"strings"
)

func FormatRoutingMessage(routeTable map[string]router.Route) string {
	var builder strings.Builder

	for _, route := range routeTable {
		fmt.Fprintf(&builder, "!%s:%d", route.DestIP, route.Metric)
	}
	return builder.String()
}
