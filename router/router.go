package router

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

type Router struct {
	IP         string
	Conn       *net.UDPConn
	RouteTable map[string]*Route
	mu         sync.Mutex
	HasChanged chan struct{}
}

type Route struct {
	DestIP      string
	Metric      int
	NextHop     string
	LastUpdated time.Time
}

func NewRouter(ip string) *Router {
	return &Router{
		IP:         ip,
		RouteTable: make(map[string]*Route),
		HasChanged: make(chan struct{}, 1),
	}
}

func (r *Router) AddRoute(destIP string, metric int, nextHop string) {
	_, exist := r.RouteTable[destIP]
	if !exist {
		r.RouteTable[destIP] = &Route{
			DestIP:      destIP,
			Metric:      metric + 1,
			NextHop:     nextHop,
			LastUpdated: time.Now(),
		}
		r.tableChange()
	}
}

func (r *Router) UpdateRoute(destIP string, metric int, nextHop string) {
	route, exist := r.RouteTable[destIP]
	if exist {
		// Se a metrica for menor, atualiza a tabela inteira
		// TODO: Se recebemos uma metrica maior, acredito que o timestamp não deve
		// Ser incrementado, já que não recebemos o real dono desse ip
		if metric <= route.Metric {
			r.RouteTable[destIP].Metric = metric
			r.RouteTable[destIP].NextHop = nextHop
			r.tableChange()
		}
	}
}

func (r *Router) RemoveRoute(destIP string) {
	delete(r.RouteTable, destIP)
}

// TODO: Tem que remover as rotas dele e que passam por ele
func (r *Router) removeInactiveRoutes() {
	for {
		// verifica a cada 10 segundos
		time.Sleep(10 * time.Second)

		r.mu.Lock()
		for destIP, route := range r.RouteTable {
			timed := time.Since(route.LastUpdated) - 35*time.Second
			fmt.Printf("timer %f \n", timed.Seconds())
			if timed > 0 {
				// Remove a rota se ela estiver inativa por mais de 35 segundos
				fmt.Printf("Removendo rota inativa: %s por %f segundos\n", destIP, timed.Seconds())
				r.RemoveRoute(destIP)
			}
		}
		r.mu.Unlock()
	}
}

func (r *Router) ToString() string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "Ip: %s\n", r.IP)

	for _, route := range r.RouteTable {
		fmt.Fprintf(&builder, "   Ip destino: %s\n", route.DestIP)
		fmt.Fprintf(&builder, "      metrica: %d\n", route.Metric)
		fmt.Fprintf(&builder, "        saida: %s\n", route.NextHop)
	}

	return builder.String()
}

func (r *Router) tableChange() {
	select {
	case r.HasChanged <- struct{}{}:
	default:
	}
}

func (r *Router) Start() {
	var (
		addr = net.UDPAddr{
			Port: 19000,
			IP:   net.ParseIP(r.IP),
		}
		err error
		wg  sync.WaitGroup
	)

	wg.Add(3)

	if r.Conn, err = net.ListenUDP("udp", &addr); err != nil {
		fmt.Printf("Listen: Error to create udp connection. Error: %s", err)
		return
	}

	go func() {
		defer wg.Done()
		r.listen()
	}()

	go func() {
		defer wg.Done()
		r.sendRouteUpdates()
	}()

	go func() {
		defer wg.Done()
		r.removeInactiveRoutes()
	}()

	wg.Wait()
}
