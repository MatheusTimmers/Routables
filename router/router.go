package router

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Router struct {
	IP         string
	RouteTable map[string]Route
}

func NewRouter(ip string) *Router {
	return &Router{
		IP:         ip,
		RouteTable: make(map[string]Route),
	}
}

func (r Router) Start() {
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

func (r Router) listen() {
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
		// TODO: Fazer alguma coisa com a mensagem
	}
}

func (r *Router) sendRouteUpdates() {
	for {
		time.Sleep(15 * time.Second)
		message := formatRoutingMessage(r.RouteTable)
		for destIP := range r.RouteTable {
			r.sendMessage(destIP, message)
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

// INFO: Falar com o herter sobre a ordem das funções impactar no code sugestions
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

func (r *Router) processMessage(message string) {
  route_table, err := parserMessageToRouteTable(message)
  if err != nil {
    fmt.Errorf("processMessage: Error to parser message")
  }

  for new_ip, new_metric := range route_table {
    // TODO: Verificar se teve novas adição, remoção ou atualização 
    for ip, metric := range r.RouteTable {
      if (new_ip == ip) {
        // mesmo destino, acho que ignora
        found := true;
      }

      if (!found) {
        r.AddRoute()
      }
    }
  }
}
