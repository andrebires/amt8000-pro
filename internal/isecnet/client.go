package isecnet

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"time"
)

const (
	cmdAuth       uint16 = 0xf0f0
	cmdDisconnect uint16 = 0xf0f1
	cmdStatus     uint16 = 0x0b4a
)

var (
	ErrInvalidAuth = errors.New("panel rejected authentication")
	ErrNoPassword  = errors.New("AMT_PASSWORD is required")
)

type Client struct {
	host     string
	port     int
	password string
	timeout  time.Duration
}

func NewClient(host string, port int, password string, timeout time.Duration) *Client {
	return &Client{host: host, port: port, password: password, timeout: timeout}
}

func (c *Client) GetStatus() (PanelStatus, error) {
	conn, err := c.connectAndAuth()
	if err != nil {
		return PanelStatus{}, err
	}
	defer conn.Close()
	defer func() {
		_ = c.writeFrame(conn, cmdDisconnect, nil)
	}()

	if err := c.writeFrame(conn, cmdStatus, nil); err != nil {
		return PanelStatus{}, err
	}
	frame, err := c.readFrame(conn)
	if err != nil {
		return PanelStatus{}, err
	}
	if len(frame.Payload) < 143 {
		return PanelStatus{}, fmt.Errorf("status payload too short: got %d want at least 143", len(frame.Payload))
	}
	return parseStatus(frame.Payload)
}

func (c *Client) connectAndAuth() (net.Conn, error) {
	if c.password == "" {
		return nil, ErrNoPassword
	}
	address := fmt.Sprintf("%s:%d", c.host, c.port)
	conn, err := net.DialTimeout("tcp", address, c.timeout)
	if err != nil {
		return nil, fmt.Errorf("connect %s: %w", address, err)
	}

	password, err := encodePassword(c.password)
	if err != nil {
		conn.Close()
		return nil, err
	}
	payload := append([]byte{0x00}, password...)
	payload = append(payload, 0x10)
	if err := c.writeFrame(conn, cmdAuth, payload); err != nil {
		conn.Close()
		return nil, err
	}
	frame, err := c.readFrame(conn)
	if err != nil {
		conn.Close()
		return nil, err
	}
	if frame.Command != cmdAuth || (len(frame.Payload) > 0 && frame.Payload[0] == 0x00) {
		conn.Close()
		return nil, ErrInvalidAuth
	}
	return conn, nil
}

func (c *Client) writeFrame(conn net.Conn, command uint16, payload []byte) error {
	if err := conn.SetWriteDeadline(time.Now().Add(c.timeout)); err != nil {
		return err
	}
	_, err := conn.Write(encodeFrame(command, payload))
	return err
}

func (c *Client) readFrame(conn net.Conn) (Frame, error) {
	if err := conn.SetReadDeadline(time.Now().Add(c.timeout)); err != nil {
		return Frame{}, err
	}
	header := make([]byte, 6)
	if _, err := io.ReadFull(conn, header); err != nil {
		return Frame{}, err
	}
	length := int(binary.BigEndian.Uint16(header[4:6]))
	if length < 2 || length > 4096 {
		return Frame{}, fmt.Errorf("invalid payload length %d", length)
	}
	rest := make([]byte, length+1)
	if _, err := io.ReadFull(conn, rest); err != nil {
		return Frame{}, err
	}
	return decodeFrame(append(header, rest...))
}
