package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"time"
)

const DefaultTimeout = 10 * time.Second

type TelnetClient interface {
	Connect() error
	io.Closer
	Send() error
	Receive() error
}

type TCPTelnetClient struct {
	address string
	timeout time.Duration
	conn    net.Conn
	in      io.ReadCloser
	out     io.Writer
}

func (t *TCPTelnetClient) Connect() error {
	conn, err := net.DialTimeout("tcp", t.address, t.timeout)
	if err != nil {
		return fmt.Errorf("error while connect to %s : %w", t.address, err)
	}
	t.conn = conn
	return nil
}

func (t *TCPTelnetClient) Close() error {
	if t.conn != nil {
		if err := t.conn.Close(); err != nil {
			return fmt.Errorf("error while closing connection to %s : %w", t.address, err)
		}
	}
	return nil
}

func (t *TCPTelnetClient) Send() error {
	_, err := io.Copy(t.conn, t.in)
	if err != nil {
		if errors.Is(err, io.EOF) {
			_, _ = fmt.Fprintln(os.Stderr, "connection closed by client (Ctrl+D)")
			return err
		}
		return fmt.Errorf("send message error: %w", err)
	}
	return nil
}

func (t *TCPTelnetClient) Receive() error {
	_, err := io.Copy(t.out, t.conn)
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("receive message error: %w", err)
	}
	return nil
}

func NewTelnetClient(address string, timeout time.Duration, in io.ReadCloser, out io.Writer) TelnetClient {
	return &TCPTelnetClient{
		address: address,
		timeout: timeout,
		in:      in,
		out:     out,
	}
}
