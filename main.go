package main

import (
	"flag"
	"fmt"
	"net"

	"routables/config"
	"routables/router"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func setupInitialUI(app *tview.Application, choiceChan chan bool) {
	// Modal para a escolha inicial
	modal := tview.NewModal().
		SetText("Deseja criar uma nova rede ou entrar em uma existente?\n\n[1] Criar Nova Rede\n[2] Entrar em Rede Existente").
		AddButtons([]string{"1: Nova Rede", "2: Entrar em Rede"}).
		SetBackgroundColor(tcell.NewRGBColor(20, 20, 40).TrueColor()).
		SetTextColor(tcell.NewRGBColor(205, 214, 244).TrueColor()).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			switch buttonIndex {
			case 0: // "0: Nova Rede"
				choiceChan <- true
			case 1: // "1: Entrar em Rede"
				choiceChan <- false
			}
			close(choiceChan) // Fecha o canal
		})

	// Configura e exibe o modal
	app.SetRoot(modal, true)
}

func detectIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		panic("Não foi possível detectar o IP da máquina")
	}

	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String()
			}
		}
	}

	panic("Não foi possível encontrar um IP válido")
}

func updateDiagram(diagramView *tview.TextView, router *router.Router) {
	diagram := router.ToString(true)
	diagramView.SetText(diagram)
}

func setupUI(app *tview.Application) (*tview.TextView, *tview.InputField, *tview.TextView, chan string) {
	commandChan := make(chan string)

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
				commandChan <- command
				fmt.Fprintf(logView, "[yellow]Comando:[-] %s\n", command)
				inputField.SetText("")
			}
		})

	currentScroll := 0
	logView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'k':
			if currentScroll > 0 {
				currentScroll--
			}
			logView.ScrollTo(currentScroll, 0)
		case 'j':
			currentScroll++
			logView.ScrollTo(currentScroll, 0)
		}
		return nil
	})

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			if app.GetFocus() == logView {
				app.SetFocus(inputField)
			} else {
				app.SetFocus(logView)
			}
			return nil
		}
		return event
	})

	// Área de diagrama
	diagramView := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(false).
		SetTextAlign(tview.AlignLeft)
	diagramView.SetBorder(true).SetTitle("Tabela de Roteamento")
	diagramView.SetBackgroundColor(tcell.NewRGBColor(20, 20, 40).TrueColor())

	horizontalFlex := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(logView, 0, 5, false).    // Logs
		AddItem(diagramView, 0, 1, false) // Diagrama da rede

	verticalFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(horizontalFlex, 0, 1, false). // Logs e diagrama
		AddItem(inputField, 1, 1, true)       // Entrada de comandos abaixo

	app.SetRoot(verticalFlex, true)

	return logView, inputField, diagramView, commandChan
}

func main() {
	// Flags
	ipFlag := flag.String("ip", "", "IP do roteador")
	configFlag := flag.String("config", "", "Arquivo de configuração dos vizinhos")
	autoScrollFlag := flag.Bool("autoscroll", true, "Ativa ou desativa o autoscroll do log (true ou false)")
	flag.Parse()

  autoScroll := *autoScrollFlag

	// Determinar o IP inicial
	var routerIP string
	if *ipFlag != "" {
		routerIP = *ipFlag
	} else {
		routerIP = detectIP()
	}

	// Determinar o arquivo de configuração
	var configFile string
	if *configFlag != "" {
		configFile = *configFlag
	} else {
		configFile = "roteadores.txt"
	}

	app := tview.NewApplication()
	choiceChan := make(chan bool)

	go func() {
		setupInitialUI(app, choiceChan)
		if err := app.Run(); err != nil {
			panic(err)
		}
	}()

	choice := <-choiceChan

	// Configuração da interface principal
	logView, _, diagramView, commandChan := setupUI(app)

	app.QueueUpdateDraw(func() {
		if choice {
			fmt.Fprintf(logView, "[green]Criando nova rede com configuração do arquivo: %s[-]\n", configFile)
		} else {
			fmt.Fprintf(logView, "[green]Entrando em rede existente usando IP: %s[-]\n", routerIP)
		}
	})

	go func() {
		router, err := config.LoadRouterConfig(configFile, routerIP)
		if err != nil {
			app.QueueUpdateDraw(func() {
				fmt.Fprintf(logView, "[red]Erro:[-] %v\n", err)
			})
			return
		}

		app.QueueUpdateDraw(func() {
			fmt.Println(logView, "[green]Sistema Iniciado:")
			updateDiagram(diagramView, router)
		})

		router.SetLogger(func(message string, isError bool) {
			app.QueueUpdateDraw(func() {
				if isError {
					fmt.Fprintf(logView, "[red]Erro:[-] %s\n", message)
				} else {
					fmt.Fprintf(logView, "%s\n", message)
				}
				updateDiagram(diagramView, router)
				if autoScroll {
					logView.ScrollToEnd()
				}
			})
		})

		router.StartCommandProcessor(commandChan)
		router.Start(choice)
	}()

	if err := app.Run(); err != nil {
		panic(err)
	}
}
