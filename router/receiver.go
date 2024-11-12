package router

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

func (r *Router) listen(addrl* net.UDPAddr) {
	connection, err := net.ListenUDP("udp", addrl)
	if err != nil {
    fmt.Printf("Listen: Error to create udp connection. Error: %s", err)
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

		fmt.Printf("Mensagem recebida de %v: \n %s\n", remoteAddr.AddrPort().Addr(), string(buf[:n]))

    // Nova mensagem, renova tabela do sender
		err = r.renewRouter(remoteAddr.AddrPort().Addr().String())
		if err != nil {
			fmt.Printf("processMessage: " + err.Error())
			continue
		}

		r.processMessage(string(buf[:n]), remoteAddr.AddrPort().Addr().String())

		fmt.Printf("Tabela de roteamento atual:\n %s\n", r.ToString())
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

// FIXME: Mutex não ficou bom
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

		if exist {
			// Se encontrou atualiza
			r.UpdateRoute(new_ip, new_metric, ip_received)
		} else {
			// Se não encontrou adiciona
			r.AddRoute(new_ip, new_metric, ip_received)
		}
	}
}

func (r *Router) renewRouter(ip_received string) error {
	_, exist := r.RouteTable[ip_received]
	if exist {
		// Renova o LastUpdate
		r.RouteTable[ip_received].LastUpdated = time.Now()
		return nil
	} else {
		// Não conhecemos o sender, retorna erro
		return fmt.Errorf("Unknown Sender \n")
	}
}
