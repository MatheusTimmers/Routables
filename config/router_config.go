package config

import (
	"fmt"
	"os"
	"routables/router"
	"strings"
)

/* func WriteRouteTable(file_name string, router router.Router) error {
  // INFO: Pode ser alterado
	file, err := os.OpenFile(file_name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)

	if err != nil {
		return fmt.Errorf("Error: Cannot open the file %s: %w", file_name, err)
	}
	defer file.Close()

	var builder strings.Builder
	for _, route := range router.RouteTable {
    if (route.Metric == 1) {
      fmt.Fprintf(&builder, "%s\n", route.DestIP)
    }
	}

	_, err = file.Write([]byte(builder.String()))
	if err != nil {
		return fmt.Errorf("Error: Cannot write in the file %s: %w", file_name, err)
	}
	return nil
}
*/

func LoadRouterConfig(fileName, ip string) (*router.Router, error) {
	// TODO: Não entendi direito como deve ficar os arquivos, revisar
	lines, err := readFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to load router config: %w", err)
	}

	r := router.NewRouter(ip)
	for _, line := range lines {
		// Carrega os roteadores vizinhos
		if line != "" {
			r.AddRoute(line, 1, line)
		}
	}
	return r, nil
}

func readFile(filename string) ([]string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(data), "\n")
	return lines, nil
}
