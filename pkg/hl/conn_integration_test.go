package hl_test

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"testing"
	"time"

	"github.com/ftl/hl-go/pkg/hl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithRigctld(t *testing.T) {
	port, err := getFreePort()
	require.NoError(t, err)
	addr := fmt.Sprintf("localhost:%d", port)

	rigctld := exec.Command("rigctld", "-m", "1", "-t", strconv.Itoa(port))
	err = rigctld.Start()
	require.NoError(t, err)

	var conn *hl.Conn
	for i := range 10 {
		_ = i
		conn, err = hl.Dial(addr)
		if err != nil {
			time.Sleep(1 * time.Second)
		} else {
			break
		}
	}
	assert.NoError(t, err)

	var response hl.Response

	response, err = conn.Execute(hl.Request{
		Command: "\\set_freq",
		Args:    []string{"14020000"},
	})
	assert.NoError(t, err)
	assert.Equal(t, response.CommandEcho, "set_freq: 14020000")
	assert.Equal(t, response.ReturnCode, hl.RigOk)

	response, err = conn.Execute(hl.Request{
		Command: "\\get_freq",
	})
	assert.NoError(t, err)
	assert.Equal(t, response.CommandEcho, "get_freq:")
	assert.Equal(t, response.Data, map[string]string{"Frequency": "14020000"})
	assert.Equal(t, response.ReturnCode, hl.RigOk)

	response, err = conn.Execute(hl.Request{
		Command: "\\get_invalid",
	})
	assert.Error(t, err)

	response, err = conn.Execute(hl.Request{
		Command: "\\get_freq",
	})
	assert.NoError(t, err)
	assert.Equal(t, response.CommandEcho, "get_freq:")
	assert.Equal(t, response.Data, map[string]string{"Frequency": "14020000"})
	assert.Equal(t, response.ReturnCode, hl.RigOk)

	err = conn.Close()
	assert.NoError(t, err)

	err = rigctld.Process.Signal(os.Interrupt)
	assert.NoError(t, err)
	rigctld.Wait()
}

func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer listener.Close()

	port := listener.Addr().(*net.TCPAddr).Port

	return port, nil
}
