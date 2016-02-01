package main

import (
	"os"
	"fmt"
	"time"
	"net"
	"strings"
	"syscall"
	"bufio"
	"golang.org/x/crypto/ssh/terminal"
)

const (
	PORT = 9034
	BATCH_SIZE = 1024
	CONN_TIMEOUT = 3
	RW_TIMEOUT = 10
)

func main() {

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: gosc <command> <ip> ... ")
		syscall.Exit(1)
	}

	fmt.Println("Executing", os.Args[1], "...")

	command := fmt.Sprintf("%s\n", os.Args[1])
	ips := make([]string, 0)

	if terminal.IsTerminal(syscall.Stdin) {
		ips = os.Args[2:]
	} else {
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Split(bufio.ScanWords)
		for scanner.Scan() {
			ips = append(ips, scanner.Text())
		}
	}

	pos := 0
	for pos < len(ips) {
		end := pos + BATCH_SIZE
		if end > len(ips) {
			end = len(ips)
		}

		responses, _ := sc(command, ips[pos:end])
		for _, r := range responses {
			fmt.Println(r)
		}

		pos = end
	}
}

func sc(command string, ips []string) ([]string, error) {
	var ch = make(chan string, len(ips))
	for _, ip := range ips {
		go func(ip string) {
			conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip, PORT), time.Second * CONN_TIMEOUT)
			if err != nil {
				ch <- fmt.Sprintf("Failed to connect to %s", ip)
				return
			}

			conn.SetDeadline(time.Now().Add(time.Second * RW_TIMEOUT))
			defer conn.Close()

			_, err = conn.Write([]byte(command))
			if err != nil {
				ch <- fmt.Sprintf("Failed to write to %s", ip)
				return
			}

			var buf = make([]byte, 1024)
			_, err = conn.Read(buf)
			if err != nil {
				ch <- fmt.Sprintf("Failed to read form %s", ip)
				return
			}

			ch <- string(buf)
		}(ip)
	}

	var responses = make([]string, 0, len(ips))
	for _ = range ips {
		r := <- ch
		responses = append(responses, strings.Trim(r, "\n "))
	}

	return responses, nil
}
