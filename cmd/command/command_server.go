package main

import (
	"log"
	"time"

	"github.com/0x42red/command_control/pkg/embeddata"
	"github.com/0x42red/command_control/pkg/server"
	"github.com/0x42red/command_control/pkg/ui"
	"github.com/jroimartin/gocui"

	gossh "golang.org/x/crypto/ssh"
)

func main() {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.SetManagerFunc(ui.Layout)
	g.Cursor = true
	go func(g *gocui.Gui) {
		for range time.Tick(1 * time.Second) {
			if g.Cursor {
				g.Cursor = false
			} else {
				g.Cursor = true
			}
		}
	}(g)

	publicKey, err := embeddata.GetPublicKey()
	if err != nil {
		log.Fatal(err)
	}

	config, err := embeddata.GetConfig()
	if err != nil {
		log.Fatal(err)
	}
	commandServer := &server.CommandServer{
		AllowedKeys: []gossh.PublicKey{
			publicKey,
		},
		Port: config.Port,
	}
	go commandServer.Start()

	err = ui.SetKeyBinding(g, commandServer)
	if err != nil {
		log.Fatal(err)
	}

	go ui.UpdateCommandServerGUI(g, commandServer)

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Fatal(err)
	}
}
