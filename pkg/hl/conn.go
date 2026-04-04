package hl

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	dialTimeout     = 5 * time.Second
	responseTimeout = 5 * time.Second

	commandPrefix     = '+'
	commandDelimiter  = '\n'
	commandEnding     = '\n'
	reportPrefix      = "RPRT "
	keyValueSeparator = ": "
)

type Conn struct {
	conn        net.Conn
	reader      *bufio.Reader
	executeLock *sync.Mutex
}

func Dial(addr string) (*Conn, error) {
	conn, err := net.DialTimeout("tcp", addr, dialTimeout)
	if err != nil {
		return nil, err
	}
	return &Conn{
		conn:        conn,
		reader:      bufio.NewReader(conn),
		executeLock: new(sync.Mutex),
	}, nil
}

func (c *Conn) Close() error {
	return c.conn.Close()
}

func (c *Conn) Execute(request Request) (Response, error) {
	c.executeLock.Lock()
	defer c.executeLock.Unlock()

	command := fmt.Sprintf("%c%s %s%c", commandPrefix, request.Command, strings.Join(request.Args, " "), commandEnding)
	_, err := fmt.Fprint(c.conn, command)
	if err != nil {
		return Response{}, fmt.Errorf("write command: %w", err)
	}

	c.conn.SetReadDeadline(time.Now().Add(responseTimeout))
	defer c.conn.SetReadDeadline(time.Time{})

	response, err := parseResponse(c.reader)
	if err == nil {
		err = ReturnCodeAsError(response.ReturnCode)
	}
	return response, err
}

func parseResponse(reader *bufio.Reader) (Response, error) {
	response := Response{Data: make(map[string]string)}
	firstLine := true
	for {
		line, err := reader.ReadString(commandDelimiter)
		if err != nil {
			return Response{}, fmt.Errorf("read response: %w", err)
		}
		line = strings.TrimRight(line, "\r\n")

		codeStr, isReportLine := strings.CutPrefix(line, reportPrefix)
		if isReportLine {
			code, err := strconv.Atoi(codeStr)
			if err != nil {
				return Response{}, fmt.Errorf("parse report: %w", err)
			}
			response.ReturnCode = ReturnCode(code)
			return response, nil
		}

		if firstLine {
			response.CommandEcho = line
			firstLine = false
			continue
		}

		sepAt := strings.Index(line, keyValueSeparator)
		if sepAt > 0 {
			key := line[:sepAt]
			value := line[sepAt+len(keyValueSeparator):]
			response.Data[key] = value
		}
	}
}
