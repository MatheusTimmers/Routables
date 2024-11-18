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
			r.log(fmt.Sprintf("Listen: Error reading udp buffer. Error: %s", err), true)
			continue
		}

		if n == 0 {
			r.log("Listen: Mensagem vazia recebida", true)
			continue
		}

		switch string(buf[:1]) {
		case "!":
			// Mensagem de tabela de roteamento
			err = r.renewRouter(remoteAddr.IP.String())
			if err != nil {
				r.log(fmt.Sprintf("renewRouter error: %s", err.Error()), true)
				continue
			}
			r.processMessage(string(buf[:n]), remoteAddr.IP.String())
			r.log(fmt.Sprintf("Mensagem processada de %v: \n %s\n", remoteAddr.IP.String(), string(buf[:n])), false)
		case "@":
			r.HandleNewNeighbor(string(buf[:n]))
		case "&":
			// Mensagem para roteamento
			split := strings.Split(string(buf[:n]), "%")
			if len(split) < 3 {
				r.log(fmt.Sprintf("Mensagem malformada: %s", string(buf[:n])), true)
				continue
			}

      r.log(fmt.Sprintf("teste %s", split[1]), false)
			if split[1] == r.IP {
        r.log(fmt.Sprintf("Mensagem recebida de %s: %s", split[0][1:], split[2]), false)
			} else {
				destIp := r.FindIp(split[1])
				if destIp != "" {
					r.log(fmt.Sprintf("Roteando mensagem para %s (de %s)", destIp, split[0]), false)
					r.sendMessage(destIp, string(buf[:n]))
				} else {
					r.log(fmt.Sprintf("Destinatário desconhecido: %s", split[1]), false)
				}
			}
		default:
			r.log(fmt.Sprintf("Mensagem com prefixo desconhecido: %s", string(buf[:1])), true)
		}

		// Limpa o buffer
		buf = make([]byte, 1024)
	}
}

// TODO: Pode ser movido para protocol/config
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
		r.log(fmt.Sprintf("processMessage: Error to parser message"), true)
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

func (r *Router) HandleNewNeighbor(message string) error {
	neighborIP := strings.TrimPrefix(message, "@")
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.RouteTable[neighborIP]; exists {
		return fmt.Errorf("Vizinho já existe na tabela de roteamento: %s", neighborIP)
	}

	r.RouteTable[neighborIP] = &Route{
		DestIP:      neighborIP,
		Metric:      1,
		NextHop:     neighborIP,
		LastUpdated: time.Now(),
	}
	r.tableChange()

	r.log(fmt.Sprintf("Novo vizinho adicionado: %s", neighborIP), false)

	return nil
}

func (r *Router) renewRouter(ip_received string) error {
	route, exist := r.RouteTable[ip_received]
	if exist && route.Metric == 1 {
		// Renova o LastUpdate
		route.LastUpdated = time.Now()
		return nil
	} else {
		// Não conhecemos o sender, retorna erro
		return fmt.Errorf("Unknown Sender %s \n", ip_received)
	}
}
