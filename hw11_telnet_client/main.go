package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var timeout time.Duration

func main() {
	flag.DurationVar(&timeout, "timeout", DefaultTimeoutSecond, "connection timeout")
	flag.Parse()

	host := flag.Arg(0)
	port := flag.Arg(1)

	address := net.JoinHostPort(host, port)
	client := NewTelnetClient(address, timeout, os.Stdin, os.Stdout)
	err := client.Connect()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to connect to %s : %v\n", address, err)
		os.Exit(1)
	}
	defer client.Close()

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := client.Send(); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "send error: %v\n", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := client.Receive(); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "receive error: %v\n", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	wg.Wait()
}
