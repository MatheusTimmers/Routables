package main

import (
	"flag"
	"fmt"
	"routables/config"
)

func main() {
	configFile := flag.String("config", "roteadores.txt", "Arquivo de configuração dos vizinhos")
	routerIP := flag.String("ip", "192.168.1.1", "IP do roteador")
	flag.Parse()

	router, err := config.LoadRouterConfig(*configFile, *routerIP)
	if err != nil {
		err = fmt.Errorf("Main: Error loading router config file: %w", err)
		panic(err)
	}

  fmt.Printf("Iniciando sistema, tabela de roteamento carrega: %s", router.ToString())

	router.Start()
}
