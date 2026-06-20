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
	ErrInvalidAuth = errors.New("invalid remote password")
	ErrNoPassword  = errors.New("remote password is required")
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
	capture, err := c.GetStatusCapture()
	if err != nil {
		return PanelStatus{}, err
	}
	return capture.Status, nil
}

func (c *Client) GetStatusCapture() (StatusCapture, error) {
	conn, err := c.connectAndAuth()
	if err != nil {
		return StatusCapture{}, err
	}
	defer conn.Close()
	defer func() {
		_ = c.writeFrame(conn, cmdDisconnect, nil)
	}()

	if err := c.writeFrame(conn, cmdStatus, nil); err != nil {
		return StatusCapture{}, err
	}
	frame, err := c.readFrame(conn)
	if err != nil {
		return StatusCapture{}, err
	}
	if len(frame.Payload) < 143 {
		return StatusCapture{}, statusPayloadTooShortError(frame)
	}
	status, err := parseStatus(frame.Payload)
	if err != nil {
		return StatusCapture{}, err
	}
	return StatusCapture{
		Status: status,
		Frame:  frame,
	}, nil
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
	if err := checkAuthResponse(frame); err != nil {
		conn.Close()
		return nil, err
	}
	return conn, nil
}

func checkAuthResponse(frame Frame) error {
	if frame.Command != cmdAuth {
		return fmt.Errorf("unexpected auth response command 0x%04x", frame.Command)
	}
	if len(frame.Payload) == 0 {
		return errors.New("empty auth response")
	}
	switch frame.Payload[0] {
	case 0x00:
		return nil
	case 0x01:
		return ErrInvalidAuth
	default:
		return fmt.Errorf("panel rejected authentication: code=0x%02x", frame.Payload[0])
	}
}

func statusPayloadTooShortError(frame Frame) error {
	if len(frame.Payload) == 1 {
		return fmt.Errorf("status payload too short: command=0x%04x code=0x%02x got 1 want at least 143", frame.Command, frame.Payload[0])
	}
	return fmt.Errorf("status payload too short: command=0x%04x got %d want at least 143 payload=% x", frame.Command, len(frame.Payload), frame.Payload)
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
