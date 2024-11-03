package router

import (
	"fmt"
	"strings"
)

type Route struct {
	DestIP   string
	Metric   int
	NextHop  string
}

func (r *Router) AddRoute(destIP string, metric int, nextHop string) {
	_, exist := r.RouteTable[destIP]
	if (!exist) {
		r.RouteTable[destIP] = Route{
			DestIP:  destIP,
			Metric:  metric,
			NextHop: nextHop,
		}
	}
}

func (r *Router) UpdateRoute(destIP string, metric int, nextHop string) {
	route, exist := r.RouteTable[destIP]
	if (!exist) || (metric < route.Metric) {
		r.RouteTable[destIP] = Route{
			DestIP:  destIP,
			Metric:  metric,
			NextHop: nextHop,
		}
	}
}

func (r *Router) RemoveRoute(destIP string) {
	delete(r.RouteTable, destIP)
}

func (r *Router) ToString() string {
	var builder strings.Builder
  fmt.Fprintf(&builder, "Ip: %s\n", r.IP)

  for _, route := range r.RouteTable {
    fmt.Fprintf(&builder, "   route: %v\n", route)
  }

  return builder.String()
}
