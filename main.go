package main

import (
	"flag"
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"routables/config"
)

func main() {
	app := tview.NewApplication()
	logView := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetChangedFunc(func() {
			app.Draw()
		})
	logView.SetBorder(true).SetTitle("Logs")
	logView.SetBackgroundColor(tcell.ColorBlack)

	inputField := tview.NewInputField()
	inputField.
		SetLabel("Comando: ").
		SetFieldWidth(30).
		SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEnter {
				command := inputField.GetText()
				fmt.Fprintf(logView, "[yellow]Comando:[-] %s\n", command)
				inputField.SetText("")
			}
		})
  inputField.SetBackgroundColor(tcell.ColorDarkBlue)

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(logView, 0, 1, false).
		AddItem(inputField, 1, 1, true)

	if err := app.SetRoot(flex, true).Run(); err != nil {
		panic(err)
	}

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
