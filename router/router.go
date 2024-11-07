package router

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

type Router struct {
	IP         string
	RouteTable map[string]Route
	mu         sync.Mutex
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
		RouteTable: make(map[string]Route),
	}
}

func (r *Router) AddRoute(destIP string, metric int, nextHop string) {
	_, exist := r.RouteTable[destIP]
	if !exist {
		r.RouteTable[destIP] = Route{
			DestIP:      destIP,
			Metric:      metric + 1,
			NextHop:     nextHop,
			LastUpdated: time.Now(),
		}
	}
}

func (r *Router) UpdateRoute(destIP string, metric int, nextHop string) {
	route, exist := r.RouteTable[destIP]
	if exist {
		// Se a metrica for menor, atualiza a tabela inteira
		// TODO: Se recebemos uma metrica maior, acredito que o timestamp não deve
		// Ser incrementado, já que não recebemos o real dono desse ip
		if metric <= route.Metric {
			r.RouteTable[destIP] = Route{
				DestIP:      destIP,
				Metric:      metric,
				NextHop:     nextHop,
				LastUpdated: time.Now(),
			}
		}
	}
}

func (r *Router) RemoveRoute(destIP string) {
	delete(r.RouteTable, destIP)
}

func (r *Router) removeStaleRoutes() {
	for {
    // verifica a cada 20 segundos
		time.Sleep(20 * time.Second)

		r.mu.Lock()
		for destIP, route := range r.RouteTable {
      time := time.Since(route.LastUpdated) - 35 * time.Second
			if time > 0 {
				// Remove a rota se ela estiver inativa por mais de 35 segundos
				fmt.Printf("Removendo rota inativa: %s por %d segundos\n", destIP, time)
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
		fmt.Fprintf(&builder, "   route: %v\n", route)
	}

	return builder.String()
}

func (r *Router) Start() {
	var wg sync.WaitGroup
	wg.Add(3)

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
		r.removeStaleRoutes()
	}()

	wg.Wait()
}
