package router

import (
	"fmt"
	"net"
	"strings"
	"time"
)

// A cada 15 segundos envia uma mensagem
// TODO: Código repetido, mover para uma função
func (r *Router) sendRouteUpdates() {
	for {
		select {
		case <-time.After(15 * time.Second):
			r.mu.Lock()

			message := formatRoutingMessage(r.RouteTable)
			for destIP := range r.RouteTable {
				r.sendMessage(destIP, message)
			}

			r.mu.Unlock()
		case <-r.HasChanged:
			r.mu.Lock()

			message := formatRoutingMessage(r.RouteTable)
			for destIP := range r.RouteTable {
				r.sendMessage(destIP, message)
			}

			r.mu.Unlock()
		}
	}
}

func (r *Router) sendMessage(destIP, message string) {
	addr := net.UDPAddr{
		Port: 19000,
		IP:   net.ParseIP(destIP),
	}
	conn, err := net.DialUDP("udp", nil, &addr)
	if err != nil {
		fmt.Printf("sendMessage: Error to connect to client %v: %v\n", destIP, err)
		return
	}
	defer conn.Close()

	_, err = conn.Write([]byte(message))
	if err != nil {
		fmt.Printf("sendMessage: Error sending a message %v: %v\n", destIP, err)
	}
}

// TODO: Gostaria que isso fosse uma função de protocol
func formatRoutingMessage(routeTable map[string]Route) string {
	var builder strings.Builder

	for _, route := range routeTable {
		fmt.Fprintf(&builder, "!%s:%d", route.DestIP, route.Metric)
	}
	return builder.String()
}
