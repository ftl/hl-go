package hl

import (
	"fmt"
)

type RigClient struct {
	conn *Conn

	addr string
}

func NewRigClient(addr string) *RigClient {
	return &RigClient{
		addr: addr,
	}
}

func (c *RigClient) Open() error {
	if c.conn != nil {
		return nil
	}

	conn, err := Dial(c.addr)
	if err != nil {
		return err
	}

	c.conn = conn

	err = c.SetVFOMode(true)
	if err != nil {
		conn.Close()
		c.conn = nil
		return fmt.Errorf("cannot enable VFO mode: %w", err)
	}

	return nil
}

func (c *RigClient) Close() error {
	if c.conn == nil {
		return nil
	}

	err := c.conn.Close()
	c.conn = nil
	return err
}

func (c *RigClient) IsConnected() bool {
	return c.conn != nil
}

func (c *RigClient) ensureConnected() error {
	if c.conn == nil {
		return fmt.Errorf("RigClient is not connected")
	}
	return nil
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
	return c.conn.Execute(request)
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
	return c.conn.ExecuteCustom(request, parseResponse)
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
	return err
}
