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

	commandPrefix            = '+'
	commandDelimiter         = '\n'
	commandEnding            = '\n'
	reportPrefix             = "RPRT "
	keyValueSeparator        = ": "
	singleValueKey           = "Value"
	singleValueLineDelimiter = "\n"
)

type ResponseParser func(*bufio.Reader) (Response, error)

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
	return c.ExecuteCustom(request, parseRegularResponse)
}

func (c *Conn) ExecuteCustom(request Request, parseResponse ResponseParser) (Response, error) {
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

func parseRegularResponse(reader *bufio.Reader) (Response, error) {
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

func parseSingleValue(reader *bufio.Reader) (Response, error) {
	response := Response{Data: make(map[string]string)}
	firstLine := true
	value := ""
	for {
		line, err := reader.ReadString(commandDelimiter)
		if err != nil {
			return Response{}, fmt.Errorf("read response: %w", err)
		}
		line = strings.TrimRight(line, "\r\n")

		returnCodeStr, isReportLine := strings.CutPrefix(line, reportPrefix)
		if isReportLine {
			returnCode, err := strconv.Atoi(returnCodeStr)
			if err != nil {
				return Response{}, fmt.Errorf("parse report: %w", err)
			}
			response.ReturnCode = ReturnCode(returnCode)
			if len(value) > 0 {
				response.Data[singleValueKey] = value
			}
			return response, nil
		}

		if firstLine {
			response.CommandEcho = line
			firstLine = false
			continue
		}

		line, returnCode, isCrippled := checkCrippledReportLine(line)

		if len(value) > 0 {
			value += singleValueLineDelimiter
		}
		value += line

		if isCrippled {
			response.ReturnCode = returnCode
			if len(value) > 0 {
				response.Data[singleValueKey] = value
			}
			return response, nil
		}
	}
}

func checkCrippledReportLine(line string) (string, ReturnCode, bool) {
	idx := strings.LastIndex(line, "RPRT ")
	if idx == -1 {
		return line, 0, false
	}

	suffix := strings.TrimSpace(line[idx+len("RPRT"):])
	returnCode, err := strconv.Atoi(suffix)
	if err != nil {
		return line, 0, false
	}

	if returnCode > 0 {
		return line, 0, false
	}

	return line[:idx], ReturnCode(returnCode), true
}
