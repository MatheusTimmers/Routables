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
  logger func(string, bool)
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
		//r.tableChange()
	}
}

func (r *Router) UpdateRoute(destIP string, metric int, nextHop string) {
	route, exist := r.RouteTable[destIP]
	if exist {
		// Se a metrica for menor, atualiza a tabela inteira
		if metric < route.Metric {
			r.RouteTable[destIP].Metric = metric
			r.RouteTable[destIP].NextHop = nextHop
			// r.tableChange()
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
      // TODO: Verificar metrica
			timed := time.Since(route.LastUpdated) - 35*time.Second
			r.log(fmt.Sprintf("timer %f \n", timed.Seconds()), false)
			if timed > 0 {
				// Remove a rota se ela estiver inativa por mais de 35 segundos
				r.log(fmt.Sprintf("Removendo rota inativa: %s por %f segundos\n", destIP, timed.Seconds()), false)
				r.RemoveRoute(destIP)
			}
		}
		r.mu.Unlock()
	}
}

func (r *Router) SetLogger(logger func(string, bool)) {
	r.logger = logger
}

func (r *Router) log(message string, isError bool) {
	if r.logger != nil {
		r.logger(message, isError)
	} else {
		fmt.Println(message)
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
    r.log(fmt.Sprintf("Listen: Error to create udp connection. Error: %s", err), true)
		return
	}

  // TODO: Criar logica de pedir para o user se vai entrar em uma rede já existente
  if false {
    destIp := "teste"
    sendStartupMessage(destIp, r)
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
