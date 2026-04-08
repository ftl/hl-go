package hl

import (
	"fmt"
)

type RigClient struct {
	conn *Conn
}

func NewRigClient(addr string) (*RigClient, error) {
	conn, err := Dial(addr)
	if err != nil {
		return nil, err
	}

	result := &RigClient{
		conn: conn,
	}

	err = result.SetVFOMode(true)
	if err != nil {
		result.Close()
		return nil, fmt.Errorf("cannot enable VFO mode: %w", err)
	}

	return result, nil
}

func (c *RigClient) Close() error {
	return c.conn.Close()
}

func (c *RigClient) get(command string, args ...string) (Response, error) {
	request := Request{
		Command: command,
		Args:    args,
	}
	return c.conn.Execute(request)
}

func (c *RigClient) getCustom(parseResponse ResponseParser, command string, args ...string) (Response, error) {
	request := Request{
		Command: command,
		Args:    args,
	}
	return c.conn.ExecuteCustom(request, parseResponse)
}

func (c *RigClient) getSingleValue(command string, args ...string) (string, error) {
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
	request := Request{
		Command: command,
		Args:    args,
	}
	_, err := c.conn.Execute(request)
	return err
}
