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

func TestConn(t *testing.T) {
	runWithRigctld(t, "Conn", func(t *testing.T, addr string) {
		var err error
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
		require.NoError(t, err)

		var response hl.Response

		response, err = conn.Execute(hl.Request{
			Command: "\\set_freq",
			Args:    []string{"14020000"},
		})
		assert.NoError(t, err)
		assert.Equal(t, "set_freq: 14020000", response.CommandEcho)
		assert.Equal(t, hl.RigOk, response.ReturnCode)

		response, err = conn.Execute(hl.Request{
			Command: "\\get_freq",
		})
		assert.NoError(t, err)
		assert.Equal(t, "get_freq:", response.CommandEcho)
		assert.Equal(t, map[string]string{"Frequency": "14020000"}, response.Data)
		assert.Equal(t, hl.RigOk, response.ReturnCode)

		response, err = conn.Execute(hl.Request{
			Command: "\\get_invalid",
		})
		assert.Error(t, err)

		response, err = conn.Execute(hl.Request{
			Command: "\\get_freq",
		})
		assert.NoError(t, err)
		assert.Equal(t, "get_freq:", response.CommandEcho)
		assert.Equal(t, map[string]string{"Frequency": "14020000"}, response.Data)
		assert.Equal(t, hl.RigOk, response.ReturnCode)

		err = conn.Close()
		assert.NoError(t, err)
	})
}

func TestRigClient(t *testing.T) {
	runWithRigctld(t, "RigClient", func(t *testing.T, addr string) {
		var err error
		var client *hl.RigClient
		for i := range 10 {
			_ = i
			client, err = hl.NewRigClient(addr)
			if err != nil {
				time.Sleep(1 * time.Second)
			} else {
				break
			}
		}
		require.NoError(t, err)

		vfoMode, err := client.CheckVFOMode()
		assert.NoError(t, err)
		assert.True(t, vfoMode)

		err = client.SetFrequency(hl.VFOA, 10102000)
		assert.NoError(t, err)

		frequency, err := client.GetFrequency(hl.VFOA)
		assert.NoError(t, err)
		assert.Equal(t, hl.Frequency(10102000), frequency)

		timestamp, err := client.GetClock()
		assert.NoError(t, err)
		assert.True(t, timestamp.IsZero())

		reply, err := client.SendRaw("1", []byte{3})
		assert.NoError(t, err)
		assert.Equal(t, "03", reply)

		bandwidths, err := client.GetModeBandwidths("CW")
		assert.NoError(t, err)
		assert.Equal(t, hl.ModeBandwidths{Mode: hl.ModeCW, Normal: 500, Narrow: 50, Wide: 2_400}, bandwidths)

		modes, err := client.GetModes()
		assert.NoError(t, err)
		assert.Equal(t, []hl.Mode{hl.ModeAM, hl.ModeCW, hl.ModeCWR, hl.ModeFM, hl.ModeLSB, hl.ModeRTTY, hl.ModeRTTYR, hl.ModeUSB, hl.ModeWFM}, hl.Modes(modes))

		clientConf, err := client.GetConf("client")
		assert.NoError(t, err)
		assert.Equal(t, "UNKNOWN", clientConf)

		functions, err := client.GetAvailableFunctions(hl.VFOA)
		assert.NoError(t, err)
		assert.Equal(t, 48, len(functions))
		assert.Equal(t, hl.AutoBandModeFunction, functions[0])
		assert.Equal(t, hl.XITFunction, functions[47])

		levels, err := client.GetAvailableLevels(hl.VFOA)
		assert.NoError(t, err)
		assert.Equal(t, 54, len(levels))
		assert.Equal(t, hl.AudioFrequencyLevel, levels[0])
		assert.Equal(t, hl.VOXGainLevel, levels[53])

		parameters, err := client.GetAvailableParameters()
		assert.NoError(t, err)
		assert.Equal(t, 14, len(parameters))
		assert.Equal(t, hl.AFIFOutputParm, parameters[0])
		assert.Equal(t, hl.TimeParm, parameters[13])

		err = client.Close()
		assert.NoError(t, err)
	})
}

func runWithRigctld(t *testing.T, name string, f func(t *testing.T, addr string)) {
	port, err := getFreePort()
	require.NoError(t, err)
	addr := fmt.Sprintf("localhost:%d", port)

	rigctld := exec.Command("rigctld", "-m", "1", "-t", strconv.Itoa(port))
	err = rigctld.Start()
	require.NoError(t, err)

	t.Run(name, func(t *testing.T) {
		f(t, addr)
	})

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
