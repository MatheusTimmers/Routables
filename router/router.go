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
	RouteTable map[string]*Route

	Conn        *net.UDPConn
	mu          sync.Mutex
	HasChanged  chan struct{}
	logger      func(string, bool)
}

type Route struct {
	DestIP      string
	Metric      int
	NextHop     string
	LastUpdated time.Time
}

func NewRouter(ip string) *Router {
	return &Router{
		IP:          ip,
		RouteTable:  make(map[string]*Route),
		HasChanged:  make(chan struct{}, 1),
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
		if metric+1 < route.Metric {
			r.RouteTable[destIP].Metric = metric
			r.RouteTable[destIP].NextHop = nextHop
			r.tableChange()
		}
	}
}

func (r *Router) RemoveRoute(destIP string) {
	for _, route := range r.RouteTable {
		if route.NextHop == destIP {
			// remove todas as rotas que esse ip ensinou
			delete(r.RouteTable, route.DestIP)
		}
	}
}

func (r *Router) removeInactiveRoutes() {
	for {
		// verifica a cada 10 segundos
		time.Sleep(10 * time.Second)

		r.mu.Lock()
		for destIP, route := range r.RouteTable {
			if route.Metric == 1 {
				timed := time.Since(route.LastUpdated) - 35*time.Second
				r.log(fmt.Sprintf("Timeout do %s. %.1f segundos \n", route.DestIP, 35+timed.Seconds()), false)
				if timed > 0 {
					// Remove a rota se ela estiver inativa por mais de 35 segundos
					r.log(fmt.Sprintf("Removendo rota inativa: %s por %f segundos\n", destIP, timed.Seconds()), false)
					r.RemoveRoute(destIP)
				}
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

func (r *Router) StartCommandProcessor(commandChan chan string) {
	go func() {
		for command := range commandChan {
			r.processCommand(command)
		}
	}()
}

func (r *Router) processCommand(command string) {
	parts := parseCommand(command)
	switch parts[0] {
	case "send":
		if len(parts) > 2 {
			r.log(fmt.Sprintf("Enviando a mensagem para %s", parts[1]), false)
			destIp := r.FindIp(parts[1])

			if destIp != "" {
				var builder strings.Builder
				fmt.Fprintf(&builder, "&%s%s%s", r.IP, "%"+parts[1], "%"+parts[2])

				r.sendMessage(destIp, builder.String())
			} else {
				r.log("Destinatário não encontrado", true)
			}
		}
	case "exit":
		r.log(fmt.Sprint("Encerrando o roteador..."), false)
	case "clear":
		r.log(fmt.Sprint("Infelizmente não deu tempo..."), false)
	default:
		r.log(fmt.Sprintf("Comando desconhecido: %s\n", command), true)
	}
}

func (r *Router) FindIp(destIp string) string {
	// Encontrar o ip dentro da tabela de roteamento
	ret := ""
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, route := range r.RouteTable {
		// procura se o ip destino esta na nossa tabela de roteamento
		if route.DestIP == destIp {
			// Se encontrou verifica se ele é vizinho direto
			if route.Metric == 1 {
				r.log("É meu vizinho, enviando a mensagem diretamente", false)
				ret = destIp
			} else {
				// Ip não é vizinhp direto, envia para o que nos ensinou sobre ele
				r.log(fmt.Sprintf("Não é meu vizinho, repassando a mensagem para o %s", route.NextHop), false)
				ret = route.NextHop
			}
			break
		}
	}
	return ret
}

func parseCommand(command string) []string {
	var parts []string
	var currentPart strings.Builder
	inQuotes := false

	for _, char := range command {
		switch char {
		case ' ':
			if inQuotes {
				currentPart.WriteRune(char)
			} else if currentPart.Len() > 0 {
				parts = append(parts, currentPart.String())
				currentPart.Reset()
			}
		case '"':
			inQuotes = !inQuotes
			if !inQuotes && currentPart.Len() > 0 {
				parts = append(parts, currentPart.String())
				currentPart.Reset()
			}
		default:
			currentPart.WriteRune(char)
		}
	}

	if currentPart.Len() > 0 {
		parts = append(parts, currentPart.String())
	}

	return parts
}

func (r *Router) GetDirectNeighbors() []string {
	neighbors := []string{}
	for _, route := range r.RouteTable {
		if route.Metric == 1 {
			neighbors = append(neighbors, route.DestIP)
		}
	}
	return neighbors
}

func (r *Router) GetNeighbors(ip string) []string {
	neighborList := []string{}

	for _, route := range r.RouteTable {
		// Se a metrica for igual a 1 significa que é vizinho direto
		// Se o NextHop for != signfica que não é vizinho do ip passado
		if route.Metric == 1 || route.NextHop != ip {
			continue
		}
		neighborList = append(neighborList, route.DestIP)
	}

	return neighborList
}

func (r *Router) ToString(colored bool) string {
	var builder strings.Builder
	// FIXME: Aqui deveria ter um lock, mas por algum motivo da DeadLock
	// Acho que é porque a cada log essa função é chamada
	//r.mu.Lock()
	//defer r.mu.Unlock()

	if colored {
		fmt.Fprintf(&builder, "[green]%s[-]\n", r.IP)
		for _, route := range r.RouteTable {
			if route.Metric == 1 {
				fmt.Fprintf(&builder, "└── [yellow]%s[-]\n", route.DestIP)
			}

			for _, neighbors := range r.GetNeighbors(route.DestIP) {
				neighbor := r.RouteTable[neighbors]
				fmt.Fprintf(&builder, "└%s [blue]%s[-]\n", strings.Repeat("──", neighbor.Metric), neighbor.DestIP)
			}
		}
	} else {
		// Deprecated
		fmt.Fprintf(&builder, "Ip: %s\n", r.IP)

		for _, route := range r.RouteTable {
			fmt.Fprintf(&builder, "   Ip destino: %s\n", route.DestIP)
			fmt.Fprintf(&builder, "      metrica: %d\n", route.Metric)
			fmt.Fprintf(&builder, "        saida: %s\n", route.NextHop)
		}
	}

	return builder.String()
}

func (r *Router) tableChange() {
	select {
	case r.HasChanged <- struct{}{}:
	default:
	}
}

func (r *Router) Start(newNetwork bool) {
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

	if !newNetwork {
		r.sendStartupMessageToAllNeighbor()
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
