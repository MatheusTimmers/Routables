package router

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
)

func (r *Router) listen() {
	addr := net.UDPAddr{
		Port: 19000,
		IP:   net.ParseIP(r.IP),
		// IP: net.ParseIP("127.0.0.1"),
	}

	connection, err := net.ListenUDP("udp", &addr)
	if err != nil {
		fmt.Printf("Listen: Error to create udp connection %s", err)
		return
	}
	defer connection.Close()

	buf := make([]byte, 1024)

	for {
		n, remoteAddr, err := connection.ReadFromUDP(buf)
		if err != nil {
			fmt.Printf("Listen: Error reading udp buffer %s", err)
			continue
		}

		fmt.Printf("Mensagem recebida de %v: %s\n", remoteAddr, string(buf[:n]))

		r.processMessage(string(buf[:n]), remoteAddr.String())
	}
}

func parserMessageToRouteTable(message string) (map[string]int, error) {
	split_msg := strings.Split(message, "!")
	route_table := make(map[string]int)

	for _, route := range split_msg {
		if route == "" {
			continue
		}

		fields := strings.Split(route, ":")
		if len(fields) != 2 {
			return nil, errors.New("parserMessageToRouteTable: Invalid format for routing table message")
		}

		ip := fields[0]
		metric, err := strconv.Atoi(fields[1])
		if err != nil {
			return nil, errors.New("parserMessageToRouteTable: Invalid format to metric in table message")
		}

		route_table[ip] = metric
	}
	return route_table, nil
}

func (r *Router) processMessage(message, ip_received string) {
	route_table, err := parserMessageToRouteTable(message)
	if err != nil {
		fmt.Printf("processMessage: Error to parser message")
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	for new_ip, new_metric := range route_table {
    _, exist := route_table[new_ip] 

		if !exist {
			// Se não encontrou adiciona com uma nova metrica
			r.AddRoute(new_ip, new_metric, ip_received)
		} else {
			// Se encontrou compara a metrica e atualiza
			// Se a metrica for a mesma, não faz nada
			if new_metric != r.RouteTable[new_ip].Metric {
				r.UpdateRoute(new_ip, new_metric, ip_received)
			}
		}
	}
}