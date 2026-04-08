package hl

import (
	"fmt"
)

type RigConnectionListener interface {
	RigConnected(connected bool)
}

type RigConnectionListenerFunc func(bool)

func (f RigConnectionListenerFunc) RigConnected(connected bool) {
	f(connected)
}

type RigClient struct {
	addr string

	conn               *Conn
	automaticReconnect bool

	listeners []any
}

func NewRigClient(addr string) *RigClient {
	return &RigClient{
		addr: addr,
	}
}

func (c *RigClient) Open(automaticReconnect bool) error {
	c.automaticReconnect = automaticReconnect
	if c.conn != nil {
		return nil
	}

	return c.connect()
}

func (c *RigClient) connect() error {
	conn, err := Dial(c.addr)
	if err != nil {
		c.handleConnectionError(err)
		return err
	}

	c.conn = conn

	automaticReconnect := c.automaticReconnect
	c.automaticReconnect = false
	err = c.SetVFOMode(true)
	c.automaticReconnect = automaticReconnect
	if err != nil {
		conn.Close()
		c.conn = nil
		return fmt.Errorf("cannot enable VFO mode: %w", err)
	}

	c.emitRigConnected(true)
	return nil
}

func (c *RigClient) Close() error {
	if c.conn == nil {
		return nil
	}

	err := c.conn.Close()
	c.conn = nil
	c.emitRigConnected(false)
	return err
}

func (c *RigClient) IsConnected() bool {
	return c.conn != nil
}

func (c *RigClient) ensureConnected() error {
	if c.conn != nil {
		return nil
	}
	if !c.automaticReconnect {
		return fmt.Errorf("RigClient is not connected")
	}

	return c.connect()
}

func (c *RigClient) handleConnectionError(err error) {
	if err != nil {
		c.conn.Close()
		c.conn = nil
		c.emitRigConnected(false)
	}
}

func (c *RigClient) get(command string, args ...string) (Response, error) {
	err := c.ensureConnected()
	if err != nil {
		return Response{}, err
	}

	request := Request{
		Command: command,
		Args:    args,
	}

	response, err := c.conn.Execute(request)
	c.handleConnectionError(err)
	return response, err
}

func (c *RigClient) getCustom(parseResponse ResponseParser, command string, args ...string) (Response, error) {
	err := c.ensureConnected()
	if err != nil {
		return Response{}, err
	}

	request := Request{
		Command: command,
		Args:    args,
	}

	response, err := c.conn.ExecuteCustom(request, parseResponse)
	c.handleConnectionError(err)
	return response, err
}

func (c *RigClient) getSingleValue(command string, args ...string) (string, error) {
	err := c.ensureConnected()
	if err != nil {
		return "", err
	}

	response, err := c.getCustom(parseSingleValue, command, args...)
	if err != nil {
		return "", err
	}

	value, err := response.GetString(singleValueKey)
	if err != nil {
		return "", err
	}

	return value, nil
}

func (c *RigClient) set(command string, args ...string) error {
	err := c.ensureConnected()
	if err != nil {
		return err
	}

	request := Request{
		Command: command,
		Args:    args,
	}

	_, err = c.conn.Execute(request)
	c.handleConnectionError(err)
	return err
}

func (c *RigClient) Notify(listener any) {
	c.listeners = append(c.listeners, listener)
}

func emit[L any](listeners []any, notify func(listener L)) {
	for i := range listeners {
		listener, ok := listeners[i].(L)
		if ok {
			notify(listener)
		}
	}
}

func (c *RigClient) emitRigConnected(connected bool) {
	emit(c.listeners, func(listener RigConnectionListener) {
		listener.RigConnected(connected)
	})
}
