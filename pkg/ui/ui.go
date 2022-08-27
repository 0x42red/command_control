package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/0x42red/reverse_shell/pkg/server"
	"github.com/jroimartin/gocui"
)

func UpdateCommandServerGUI(g *gocui.Gui, cs *server.CommandServer) {
	for {
		select {
		case <-cs.Update:
			g.Update(func(g *gocui.Gui) error {
				v, err := g.View("console")
				if err != nil {
					return err
				}
				v.Clear()
				v.SetCursor(0, 0)
				v.SetOrigin(0, 0)
				for _, line := range cs.CurrentClient.GetBuffer() {
					fmt.Fprintln(v, line)
				}

				return nil
			})

			g.Update(func(g *gocui.Gui) error {
				v, err := g.View("client_list")
				if err != nil {
					return err
				}
				v.Clear()
				v.SetCursor(0, 0)
				for _, client := range cs.Clients {
					if client == cs.CurrentClient {
						fmt.Fprintln(v, "\033[32m>> ", client.RemoteAddr, "\033[0m")
					} else {
						fmt.Fprintln(v, client.RemoteAddr)
					}
				}
				return nil
			})
		default:
			if cs.CurrentClient == nil {
				g.Update(func(g *gocui.Gui) error {
					v, err := g.View("console")
					if err != nil {
						return err
					}

					fmt.Fprint(v, ".")
					return nil
				})
			}
			time.Sleep(500 * time.Millisecond)
		}
	}
}

func Layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView("console", 0, 0, (maxX/3)*2, maxY-5); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Autoscroll = true
		v.Wrap = true
		v.Title = "Console"
		fmt.Fprintln(v, "   ▓▓▓▓▓▓  ▓▓   ▓▓ ▓▓   ▓▓ ▓▓▓▓▓▓         \033[31m▓▓▓▓▓░   ▓▓▓▓▒░  ▓▓▓▓▓▒\033[0m")
		fmt.Fprintln(v, "  ▓▓  ▓▓▓▓  ▓▓ ▓▓  ▓▓   ▓▓      ▓▓        \033[31m▓▒   ▓▒  ▓▒      ▓▒   ▓▒\033[0m")
		fmt.Fprintln(v, "  ▓▓ ▓▓ ▓▓   ▓▓▓   ▓▓▓▓▓▓▓  ▓▓▓▓▓         \033[31m▓▓▓▓▓░   ▓▓▒░    ▓░   ▓▒░\033[0m")
		fmt.Fprintln(v, "  ▓▓▓▓  ▓▓  ▓▓ ▓▓       ▓▓ ▓▓             \033[31m▓▒   ▓░  ▓▒      ▓▒   ▓░\033[0m")
		fmt.Fprintln(v, "   ▓▓▓▓▓▓  ▓▓   ▓▓      ▓▓ ▓▓▓▓▓▓▓   ░░   \033[31m▓░   ▓░  ▓▓▓▓▒░  ▓▓▓▓\033[0mCommandCenter")
		fmt.Fprintln(v, "")
		fmt.Fprintln(v, "Usage:")
		fmt.Fprintln(v, "   - Have multiple clients connect to this machine via SSH")
		fmt.Fprintln(v, "   - Type commands and hit enter to send to client")
		fmt.Fprintln(v, "   - UP and DOWN arrows cycle through connected client")
		fmt.Fprintln(v, "=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=")
		fmt.Fprintln(v, "")
		fmt.Fprint(v, "\033[33mNo clients\033[0m")
	}

	if v, err := g.SetView("client_list", (maxX/3)*2+1, 0, maxX-1, maxY-5); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Connections"
		fmt.Fprintln(v, "\033[33mAwaiting connections\033[0m")
	}

	if v, err := g.SetView("command", 0, maxY-4, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		v.Editable = true
		v.Wrap = true
		v.Highlight = true
		if _, err := g.SetCurrentView("command"); err != nil {
			return err
		}
	}

	// if v, err := g.SetView("hello", maxX/2-7, maxY/2, maxX/2+7, maxY/2+2); err != nil {
	// 	v.
	// 	if err != gocui.ErrUnknownView {
	// 		return err
	// 	}

	// 	v.BgColor = gocui.ColorGreen
	// 	fmt.Fprintln(v, "Connected")
	// }

	return nil
}

func SetKeyBinding(g *gocui.Gui, commandServer *server.CommandServer) error {
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		return err
	}

	if err := g.SetKeybinding("command", gocui.KeyEnter, gocui.ModNone, enterCommand(commandServer)); err != nil {
		return err
	}

	if err := g.SetKeybinding("", gocui.KeyArrowDown, gocui.ModNone, changeClientCommand(commandServer, false)); err != nil {
		return err
	}
	if err := g.SetKeybinding("", gocui.KeyArrowUp, gocui.ModNone, changeClientCommand(commandServer, true)); err != nil {
		return err
	}

	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func enterCommand(cs *server.CommandServer) func(*gocui.Gui, *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		if cs.CurrentClient == nil {
			return nil
		}
		cs.CurrentClient.AddCommands(strings.TrimSpace(string(v.Buffer())))
		cs.Update <- true

		if _, err := g.SetCurrentView("command"); err != nil {
			return err
		}
		v.Clear()
		v.SetCursor(0, 0)
		return nil
	}
}

func changeClientCommand(cs *server.CommandServer, up bool) func(*gocui.Gui, *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		if cs.CurrentClient == nil {
			return nil
		}
		currentClientIndex := 0
		for i, client := range cs.Clients {
			if client == cs.CurrentClient {
				currentClientIndex = i
				break
			}
		}

		modifier := 1
		if up {
			modifier = -1
		}
		currentClientIndex += modifier

		// If we are out of bounds wrap it
		if currentClientIndex > len(cs.Clients)-1 {
			currentClientIndex = 0
		}
		if currentClientIndex < 0 {
			currentClientIndex = len(cs.Clients) - 1
		}

		cs.CurrentClient = cs.Clients[currentClientIndex]
		cs.Update <- true

		return nil
	}
}
