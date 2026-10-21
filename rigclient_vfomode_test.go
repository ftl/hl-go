package hl_test

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ftl/hl-go"
)

func TestRigClient_OpenSkipsSetVFOOptIfVFOModeIsAlreadyEnabled(t *testing.T) {
	server := newFakeRigctld(t, true)
	defer server.close()

	client := hl.NewRigClient(server.addr())
	err := client.Open(false)
	require.NoError(t, err)
	defer client.Close()

	assert.Equal(t, []string{"\\chk_vfo"}, server.receivedCommands())
}

func TestRigClient_OpenSendsSetVFOOptIfVFOModeIsDisabled(t *testing.T) {
	server := newFakeRigctld(t, false)
	defer server.close()

	client := hl.NewRigClient(server.addr())
	err := client.Open(false)
	require.NoError(t, err)
	defer client.Close()

	assert.Equal(t, []string{"\\chk_vfo", "\\set_vfo_opt"}, server.receivedCommands())
}

type fakeRigctld struct {
	listener      net.Listener
	vfoModeOption string

	lock     *sync.Mutex
	commands []string
}

func newFakeRigctld(t *testing.T, vfoModeEnabled bool) *fakeRigctld {
	t.Helper()

	listener, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)

	result := &fakeRigctld{
		listener:      listener,
		vfoModeOption: boolToOption(vfoModeEnabled),
		lock:          new(sync.Mutex),
	}
	go result.run()

	return result
}

func boolToOption(value bool) string {
	if value {
		return "1"
	}
	return "0"
}

func (s *fakeRigctld) addr() string {
	return s.listener.Addr().String()
}

func (s *fakeRigctld) close() {
	s.listener.Close()
}

func (s *fakeRigctld) receivedCommands() []string {
	s.lock.Lock()
	defer s.lock.Unlock()
	return append([]string{}, s.commands...)
}

func (s *fakeRigctld) run() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return
		}
		go s.handle(conn)
	}
}

func (s *fakeRigctld) handle(conn net.Conn) {
	defer conn.Close()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		fields := strings.Fields(strings.TrimPrefix(scanner.Text(), "+"))
		if len(fields) == 0 {
			continue
		}
		command := fields[0]

		s.lock.Lock()
		s.commands = append(s.commands, command)
		s.lock.Unlock()

		if command == "\\chk_vfo" {
			fmt.Fprintf(conn, "ChkVFO: %s\n", s.vfoModeOption)
			continue
		}
		fmt.Fprintf(conn, "%s: %s\nRPRT 0\n", strings.TrimPrefix(command, "\\"), strings.Join(fields[1:], " "))
	}
}
