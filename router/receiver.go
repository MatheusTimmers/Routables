package router

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func (r *Router) listen() {
  defer r.Conn.Close()
	buf := make([]byte, 1024)

	for {
		n, remoteAddr, err := r.Conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Printf("Listen: Error reading udp buffer %s", err)
			continue
		}

		fmt.Printf("Mensagem recebida de %v: \n %s\n", remoteAddr.IP.String(), string(buf[:n]))

    // Nova mensagem, renova tabela do sender
		err = r.renewRouter(remoteAddr.IP.String())
		if err != nil {
			fmt.Printf("processMessage: " + err.Error())
			continue
		}

		r.processMessage(string(buf[:n]), remoteAddr.IP.String())

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
    if new_ip == r.IP {
      continue
    }
		_, exist := r.RouteTable[new_ip]

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
