package hl

import "fmt"

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

func (c *RigClient) set(command string, args ...string) error {
	request := Request{
		Command: command,
		Args:    args,
	}
	_, err := c.conn.Execute(request)
	return err
}

func (c *RigClient) SetVFOMode(enabled bool) error {
	return c.set("\\set_vfo_opt", boolToHL(enabled))
}

func (c *RigClient) GetFrequency(vfo VFO) (Frequency, error) {
	response, err := c.get("\\get_freq", string(vfo))
	if err != nil {
		return 0, err
	}

	frequency, err := response.GetFloat64("Frequency")
	if err != nil {
		return 0, err
	}

	return Frequency(frequency), nil
}

func (c *RigClient) SetFrequency(vfo VFO, frequency Frequency) error {
	return c.set("\\set_freq", string(vfo), frequencyToHL(frequency))
}
