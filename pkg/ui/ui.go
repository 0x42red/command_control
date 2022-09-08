package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/0x42red/command_control/pkg/server"
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

				if !v.Autoscroll {
					ox, oy := v.Origin()
					v.SetOrigin(ox, oy+cs.CurrentClient.CursorY)
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
					select {
					case <-client.Context.Done():
						if client == cs.CurrentClient {
							fmt.Fprintln(v, "\033[31m>> ", client.RemoteAddr, "\033[0m")
						} else {
							fmt.Fprintln(v, "\033[31m", client.RemoteAddr, "\033[0m")
						}
					default:
						if client == cs.CurrentClient {
							fmt.Fprintln(v, "\033[32m>> ", client.RemoteAddr, "\033[0m")
						} else {
							fmt.Fprintln(v, client.RemoteAddr)
						}
					}
				}

				// Great area to debug things with
				// console, _ := g.View("console")
				// _, oy := console.Origin()
				// _, y := console.Size()
				// fmt.Fprintln(v, "ViewBufferLines:", len(console.ViewBufferLines()))
				// fmt.Fprintln(v, "Bufferlines:", len(console.BufferLines()))
				// fmt.Fprintln(v, "oy+cursorY:", oy+cs.CurrentClient.CursorY)
				// fmt.Fprintln(v, "console size:", fmt.Sprint(console.Size()))
				// fmt.Fprintln(v, "console origin X, Y:", fmt.Sprint(console.Origin()))
				// fmt.Fprintln(v, "UserCursorPosition:", cs.CurrentClient.CursorY)
				// fmt.Fprintln(v, "UserCursorScrolling:", cs.CurrentClient.Scrolling)
				// fmt.Fprintln(v, "Logic:", fmt.Sprint(oy+cs.CurrentClient.CursorY > len(console.ViewBufferLines())-y-1))
				// fmt.Fprintln(v, "Autoscrolling:", console.Autoscroll)
				// fmt.Fprintln(v, "Break Down:", fmt.Sprint(len(console.ViewBufferLines())-y-1))
				// fmt.Fprintln(v, "Simple:", oy+cs.CurrentClient.CursorY)
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
		fmt.Fprintln(v, "   ▓▓▓▓▓▓  ▓▓   ▓▓      ▓▓ ▓▓▓▓▓▓▓   ░░   \033[31m▓░   ▓░  ▓▓▓▓▒░  ▓▓▓▓\033[0mCommandControl")
		fmt.Fprintln(v, "")
		fmt.Fprintln(v, "\033[31mDisclaimer:\033[0m")
		fmt.Fprintln(v, "    Only use this software on systems you have permission to exfil.  Don't be a goof")
		fmt.Fprintln(v, "")
		fmt.Fprintln(v, "Usage:")
		fmt.Fprintln(v, "   - Have multiple clients connect to this machine via client binary")
		fmt.Fprintln(v, "   - Type commands and hit enter to send to selected client")
		fmt.Fprintln(v, "   - UP and DOWN arrows cycle through connected clients")
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

	if err := g.SetKeybinding("", gocui.KeyCtrlI, gocui.ModNone, scrollCommand(commandServer, true)); err != nil {
		return err
	}
	if err := g.SetKeybinding("", gocui.KeyCtrlK, gocui.ModNone, scrollCommand(commandServer, false)); err != nil {
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

func scrollCommand(cs *server.CommandServer, up bool) func(*gocui.Gui, *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		if cs.CurrentClient == nil {
			return nil
		}

		console, _ := g.View("console")
		_, y := console.Size()
		_, oy := console.Origin()
		if len(console.ViewBufferLines()) < y || (!cs.CurrentClient.Scrolling && !up) {
			return nil
		}

		if cs.CurrentClient.Scrolling && oy > len(console.ViewBufferLines())-y-1 {
			console.Autoscroll = true
			cs.CurrentClient.Scrolling = false

			return nil
		}

		if !cs.CurrentClient.Scrolling {
			cs.CurrentClient.CursorY = len(console.ViewBufferLines()) - y
		}

		dy := 1
		if up {
			dy = -1
		}

		console.Autoscroll = false
		cs.CurrentClient.Scrolling = true
		cs.CurrentClient.CursorY += dy
		if cs.CurrentClient.CursorY < 0 {
			cs.CurrentClient.CursorY = 0
		}
		cs.Update <- true
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
