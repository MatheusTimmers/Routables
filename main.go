package main

import (
	"flag"
	"fmt"
	"strings"

	"routables/config"
  "routables/router"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func setupUI() (*tview.Application, *tview.TextView, *tview.InputField, *tview.TextView) {
	app := tview.NewApplication()

	// Área de log
	logView := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetTextColor(tcell.NewRGBColor(205, 214, 244).TrueColor()).
		SetChangedFunc(func() {
			app.Draw()
		})
	logView.SetBorder(true).SetTitle("Logs")
	logView.SetBackgroundColor(tcell.NewRGBColor(20, 20, 40).TrueColor())

	// Área de entrada
	inputField := tview.NewInputField()
	inputField.SetLabel("Comando: ").
		SetFieldWidth(0).
		SetFieldTextColor(tcell.NewRGBColor(30, 30, 46).TrueColor()).
		SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEnter {
				command := inputField.GetText()
				fmt.Fprintf(logView, "[yellow]Comando:[-] %s\n", command)
				inputField.SetText("")
			}
		})

	// Área de diagrama
	diagramView := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(false).
		SetTextAlign(tview.AlignCenter)
	diagramView.SetBorder(true).SetTitle("Diagrama da Rede")
	diagramView.SetBackgroundColor(tcell.NewRGBColor(20, 20, 40).TrueColor())

	horizontalFlex := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(logView, 0, 5, false).    // Logs
		AddItem(diagramView, 0, 1, false) // Diagrama da rede

	verticalFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(horizontalFlex, 0, 1, false). // Logs e diagrama
		AddItem(inputField, 1, 1, true)       // Entrada de comandos abaixo

	app.SetRoot(verticalFlex, true)

	return app, logView, inputField, diagramView
}

func updateDiagram(diagramView *tview.TextView, router *router.Router) {
	diagram := buildDiagram(router)
	diagramView.SetText(diagram)
}

func buildDiagram(router *router.Router) string {
	var builder strings.Builder

	fmt.Fprintf(&builder, "[green]%s[-]\n", router.IP)

	for destIP := range router.RouteTable {
		fmt.Fprintf(&builder, "  └── [yellow]%s[-]\n", destIP)
	}

	return builder.String()
}

func main() {
	app, logView, _, diagramView := setupUI()

	configFile := flag.String("config", "roteadores.txt", "Arquivo de configuração dos vizinhos")
	routerIP := flag.String("ip", "192.168.1.1", "IP do roteador")
	flag.Parse()

	go func() {
		router, err := config.LoadRouterConfig(*configFile, *routerIP)
		if err != nil {
			app.QueueUpdateDraw(func() {
				fmt.Fprintf(logView, "[red]Erro:[-] %v\n", err)
			})
			return
		}

		app.QueueUpdateDraw(func() {
			fmt.Fprintf(logView, "[green]Sistema Iniciado:[-] %s\n", router.ToString())
			updateDiagram(diagramView, router) // Atualiza o diagrama inicial
		})

		router.SetLogger(func(message string, isError bool) {
			app.QueueUpdateDraw(func() {
				if isError {
					fmt.Fprintf(logView, "[red]Erro:[-] %s\n", message)
				} else {
					fmt.Fprintf(logView, "%s\n", message)
				}
				updateDiagram(diagramView, router) // Atualiza o diagrama ao processar logs
			})
		})

		router.Start()
	}()

	if err := app.Run(); err != nil {
		panic(err)
	}
}
