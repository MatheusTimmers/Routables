package router

import (
	"fmt"
	"strings"
	"sync"
)

type Router struct {
	IP         string
	RouteTable map[string]Route
	mu         sync.Mutex
}

type Route struct {
	DestIP  string
	Metric  int
	NextHop string
}

func NewRouter(ip string) *Router {
	return &Router{
		IP:         ip,
		RouteTable: make(map[string]Route),
	}
}

func (r *Router) AddRoute(destIP string, metric int, nextHop string) {
	_, exist := r.RouteTable[destIP]
	if !exist {
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

func (r *Router) Start() {
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		r.listen()
	}()

	go func() {
		defer wg.Done()
		r.sendRouteUpdates()
	}()

	wg.Wait()
}
