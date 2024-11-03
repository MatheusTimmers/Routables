package main

import (
	"fmt"
	"routables/config"
)

func main() {
  r, err := config.LoadRouterConfig("roteadores.txt", "192.168.1.1")
  if (err != nil) {
    err = fmt.Errorf("Main: Error loading router config file: %w", err)
    panic(err)
  }

  fmt.Printf(r.ToString())
}
