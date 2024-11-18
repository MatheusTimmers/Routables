package router

import (
	"fmt"
	"net"
	"strings"
	"time"
)

// A cada 15 segundos envia uma mensagem
func (r *Router) sendRouteUpdates() {
	for {
		select {
		case <-time.After(15 * time.Second):
			r.innerSendRouteUpdates()
		case <-r.HasChanged:
			r.innerSendRouteUpdates()
		}
	}
}

func (r *Router) innerSendRouteUpdates() {
	r.mu.Lock()
	message := formatRoutingMessage(r.RouteTable)
	for _, route := range r.RouteTable {
		if route.Metric == 1 {
			r.sendMessage(route.DestIP, message)
		}
	}
	r.mu.Unlock()
}

func (r *Router) sendMessage(destIP, message string) {
	addr := net.UDPAddr{
		Port: 19000,
		IP:   net.ParseIP(destIP),
	}

	_, err := r.Conn.WriteToUDP([]byte(message), &addr)
	if err != nil {
		r.log(fmt.Sprintf("sendMessage: Error to connect to client %v: %v\n", destIP, err), true)
		return
	}
}

func (r *Router) sendStartupMessageToAllNeighbor() {
	var builder strings.Builder
	fmt.Fprintf(&builder, "@%s", r.IP)

	r.mu.Lock()
	defer r.mu.Unlock()

	for _, route := range r.RouteTable {
		if route.Metric == 1 {
			r.sendMessage(route.DestIP, builder.String())
		}
	}
}

// INFO: Gostaria de ter criado um arquivo protocol.go, e que tivesse todos os parses
// de mensagens lá, mas não consegui fazer ;-;
func formatRoutingMessage(routeTable map[string]*Route) string {
	var builder strings.Builder

	for _, route := range routeTable {
		fmt.Fprintf(&builder, "!%s:%d", route.DestIP, route.Metric)
	}
	return builder.String()
}
