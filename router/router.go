package router

import (
	"fmt"
	"net"
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
		//	IP:   net.ParseIP(r.IP),
		IP: net.ParseIP("0.0.0.0"),
	}

	connection, err := net.ListenUDP("udp", &addr)
	if err != nil {
		fmt.Printf("Listen: Error to create udp connection %w", err)
		return
	}
	defer connection.Close()

	buf := make([]byte, 1024)

	for {
		n, remoteAddr, err := connection.ReadFromUDP(buf)
		if err != nil {
			fmt.Printf("Listen: Error reading udp buffer %w", err)
			continue
		}

		fmt.Printf("Mensagem recebida de %v: %s\n", remoteAddr, string(buf[:n]))
		// TODO: Fazer alguma coisa com a mensagem
	}
}

func (r *Router) sendRouteUpdates() {
	for {
		time.Sleep(15 * time.Second)
		message := r.formatMessage()
		for destIP := range r.RouteTable {
			r.sendMessage(destIP, message)
		}
	}
}

func (r *Router) formatMessage() string {
	return r.ToString()
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
