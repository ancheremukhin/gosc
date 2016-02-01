package main

import "os"
import "fmt"
import "time"
import "net"
import "strings"

const (
	PORT = 9034
	BATCH_SIZE = 512
	CONN_TIMEOUT = 3
	RW_TIMEOUT = 10
)

func main() {

	fmt.Println("Executing", os.Args[1], "...")
	command := fmt.Sprintf("%s\n", os.Args[1])

	pos := 2
	for pos < len(os.Args) {
		end := pos + BATCH_SIZE
		if end > len(os.Args) {
			end = len(os.Args)
		}

		responses, _ := sc(command, os.Args[pos:end])
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
	for range ips {
		r := <- ch
		responses = append(responses, strings.Trim(r, "\n "))
	}

	return responses, nil
}
