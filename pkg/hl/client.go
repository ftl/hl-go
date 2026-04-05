package hl

import (
	"bufio"
	"fmt"
	"strings"
	"time"
)

const timeFormat = "2006-01-02T15:04:05.999Z07:00"

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

func (c *RigClient) GetMode(vfo VFO) (string, int, error) {
	response, err := c.get("\\get_mode", string(vfo))
	if err != nil {
		return "", 0, err
	}

	mode, err := response.GetString("Mode")
	if err != nil {
		return "", 0, err
	}

	passband, err := response.GetInt("Passband")
	if err != nil {
		return "", 0, err
	}

	return mode, passband, nil
}

func (c *RigClient) SetMode(vfo VFO, mode string, passband int) error {
	return c.set("\\set_mode", string(vfo), mode, fmt.Sprintf("%d", passband))
}

func (c *RigClient) GetVFO() (VFO, error) {
	response, err := c.get("\\get_vfo")
	if err != nil {
		return "", err
	}

	vfo, err := response.GetString("VFO")
	if err != nil {
		return "", err
	}

	return VFO(vfo), nil
}

func (c *RigClient) SetVFO(vfo VFO) error {
	return c.set("\\set_vfo", string(vfo))
}

func (c *RigClient) GetRIT(vfo VFO) (int, error) {
	response, err := c.get("\\get_rit", string(vfo))
	if err != nil {
		return 0, err
	}

	rit, err := response.GetInt("RIT")
	if err != nil {
		return 0, err
	}

	return rit, nil
}

func (c *RigClient) SetRIT(vfo VFO, rit int) error {
	return c.set("\\set_rit", string(vfo), fmt.Sprintf("%d", rit))
}

func (c *RigClient) GetXIT(vfo VFO) (int, error) {
	response, err := c.get("\\get_xit", string(vfo))
	if err != nil {
		return 0, err
	}

	xit, err := response.GetInt("XIT")
	if err != nil {
		return 0, err
	}

	return xit, nil
}

func (c *RigClient) SetXIT(vfo VFO, xit int) error {
	return c.set("\\set_xit", string(vfo), fmt.Sprintf("%d", xit))
}

func (c *RigClient) GetPTT(vfo VFO) (int, error) {
	response, err := c.get("\\get_ptt", string(vfo))
	if err != nil {
		return 0, err
	}

	ptt, err := response.GetInt("PTT")
	if err != nil {
		return 0, err
	}

	return ptt, nil
}

func (c *RigClient) SetPTT(vfo VFO, ptt int) error {
	return c.set("\\set_ptt", string(vfo), fmt.Sprintf("%d", ptt))
}

func (c *RigClient) GetSplitVFO(vfo VFO) (int, VFO, error) {
	response, err := c.get("\\get_split_vfo", string(vfo))
	if err != nil {
		return 0, "", err
	}

	split, err := response.GetInt("Split")
	if err != nil {
		return 0, "", err
	}

	txVFO, err := response.GetString("TX VFO")
	if err != nil {
		return 0, "", err
	}

	return split, VFO(txVFO), nil
}

func (c *RigClient) SetSplitVFO(vfo VFO, split int, txVFO VFO) error {
	return c.set("\\set_split_vfo", string(vfo), fmt.Sprintf("%d", split), string(txVFO))
}

func (c *RigClient) GetSplitFrequency(vfo VFO) (Frequency, error) {
	response, err := c.get("\\get_split_freq", string(vfo))
	if err != nil {
		return 0, err
	}

	frequency, err := response.GetFloat64("TX Frequency")
	if err != nil {
		return 0, err
	}

	return Frequency(frequency), nil
}

func (c *RigClient) SetSplitFrequency(vfo VFO, txFrequency Frequency) error {
	return c.set("\\set_split_freq", string(vfo), frequencyToHL(txFrequency))
}

func (c *RigClient) GetSplitMode(vfo VFO) (string, int, error) {
	response, err := c.get("\\get_split_mode", string(vfo))
	if err != nil {
		return "", 0, err
	}

	mode, err := response.GetString("TX Mode")
	if err != nil {
		return "", 0, err
	}

	passband, err := response.GetInt("TX Passband")
	if err != nil {
		return "", 0, err
	}

	return mode, passband, nil
}

func (c *RigClient) SetSplitMode(vfo VFO, txMode string, txPassband int) error {
	return c.set("\\set_split_mode", string(vfo), txMode, fmt.Sprintf("%d", txPassband))
}

func (c *RigClient) GetSplitFreqMode(vfo VFO) (Frequency, string, int, error) {
	response, err := c.get("\\get_split_freq_mode", string(vfo))
	if err != nil {
		return 0, "", 0, err
	}

	freq, err := response.GetFloat64("TX Frequency")
	if err != nil {
		return 0, "", 0, err
	}

	mode, err := response.GetString("TX Mode")
	if err != nil {
		return 0, "", 0, err
	}

	passband, err := response.GetInt("TX Passband")
	if err != nil {
		return 0, "", 0, err
	}

	return Frequency(freq), mode, passband, nil
}

func (c *RigClient) SetSplitFreqMode(vfo VFO, txFrequency Frequency, txMode string, txPassband int) error {
	return c.set("\\set_split_freq_mode", string(vfo), frequencyToHL(txFrequency), txMode, fmt.Sprintf("%d", txPassband))
}

func (c *RigClient) GetAntenna(vfo VFO, antCurr int) (int, int, int, int, error) {
	response, err := c.get("\\get_ant", string(vfo), fmt.Sprintf("%d", antCurr))
	if err != nil {
		return 0, 0, 0, 0, err
	}

	ant, err := response.GetInt("AntCurr")
	if err != nil {
		return 0, 0, 0, 0, err
	}

	option, err := response.GetInt("Option")
	if err != nil {
		return 0, 0, 0, 0, err
	}

	antTx, err := response.GetInt("AntTx")
	if err != nil {
		return 0, 0, 0, 0, err
	}

	antRx, err := response.GetInt("AntRx")
	if err != nil {
		return 0, 0, 0, 0, err
	}

	return ant, option, antTx, antRx, nil
}

func (c *RigClient) SetAntenna(vfo VFO, antenna int, option int) error {
	return c.set("\\set_ant", string(vfo), fmt.Sprintf("%d", antenna), fmt.Sprintf("%d", option))
}

func (c *RigClient) GetFunc(vfo VFO, funcName string) (int, error) {
	response, err := c.getCustom(parseSingleValue, "\\get_func", string(vfo), funcName)
	if err != nil {
		return 0, err
	}

	status, err := response.GetInt(singleValueKey)
	if err != nil {
		return 0, err
	}

	return status, nil
}

func (c *RigClient) SetFunc(vfo VFO, funcName string, status int) error {
	return c.set("\\set_func", string(vfo), funcName, fmt.Sprintf("%d", status))
}

func (c *RigClient) GetLevel(vfo VFO, levelName string) (float64, error) {
	response, err := c.getCustom(parseSingleValue, "\\get_level", string(vfo), levelName)
	if err != nil {
		return 0, err
	}

	value, err := response.GetFloat64(singleValueKey)
	if err != nil {
		return 0, err
	}

	return value, nil
}

func (c *RigClient) SetLevel(vfo VFO, levelName string, value float64) error {
	return c.set("\\set_level", string(vfo), levelName, fmt.Sprintf("%f", value))
}

func (c *RigClient) GetParm(parmName string) (string, error) {
	response, err := c.getCustom(parseSingleValue, "\\get_parm", parmName)
	if err != nil {
		return "", err
	}

	value, err := response.GetString(singleValueKey)
	if err != nil {
		return "", err
	}

	return value, nil
}

func (c *RigClient) SetParm(parmName string, value string) error {
	return c.set("\\set_parm", parmName, value)
}

func (c *RigClient) GetMemory(vfo VFO) (int, error) {
	response, err := c.get("\\get_mem", string(vfo))
	if err != nil {
		return 0, err
	}

	memory, err := response.GetInt("Memory#")
	if err != nil {
		return 0, err
	}

	return memory, nil
}

func (c *RigClient) SetMemory(vfo VFO, memory int) error {
	return c.set("\\set_mem", string(vfo), fmt.Sprintf("%d", memory))
}

func (c *RigClient) SetBank(vfo VFO, bank int) error {
	return c.set("\\set_bank", string(vfo), fmt.Sprintf("%d", bank))
}

func (c *RigClient) VFOOp(vfo VFO, op string) error {
	return c.set("\\vfo_op", string(vfo), op)
}

func (c *RigClient) Scan(vfo VFO, scanFct string, scanChannel int) error {
	return c.set("\\scan", string(vfo), scanFct, fmt.Sprintf("%d", scanChannel))
}

func (c *RigClient) GetRepeaterShift(vfo VFO) (string, error) {
	response, err := c.get("\\get_rptr_shift", string(vfo))
	if err != nil {
		return "", err
	}

	shift, err := response.GetString("Rptr Shift")
	if err != nil {
		return "", err
	}

	return shift, nil
}

func (c *RigClient) SetRepeaterShift(vfo VFO, shift string) error {
	return c.set("\\set_rptr_shift", string(vfo), shift)
}

func (c *RigClient) GetRepeaterOffset(vfo VFO) (int, error) {
	response, err := c.get("\\get_rptr_offs", string(vfo))
	if err != nil {
		return 0, err
	}

	offset, err := response.GetInt("Rptr Offset")
	if err != nil {
		return 0, err
	}

	return offset, nil
}

func (c *RigClient) SetRepeaterOffset(vfo VFO, offset int) error {
	return c.set("\\set_rptr_offs", string(vfo), fmt.Sprintf("%d", offset))
}

func (c *RigClient) GetCTCSSCode(vfo VFO) (int, error) {
	response, err := c.get("\\get_ctcss_tone", string(vfo))
	if err != nil {
		return 0, err
	}

	code, err := response.GetInt("CTCSS Tone")
	if err != nil {
		return 0, err
	}

	return code, nil
}

func (c *RigClient) SetCTCSSCode(vfo VFO, code int) error {
	return c.set("\\set_ctcss_tone", string(vfo), fmt.Sprintf("%d", code))
}

func (c *RigClient) GetDCSCode(vfo VFO) (int, error) {
	response, err := c.get("\\get_dcs_code", string(vfo))
	if err != nil {
		return 0, err
	}

	code, err := response.GetInt("DCS Code")
	if err != nil {
		return 0, err
	}

	return code, nil
}

func (c *RigClient) SetDCSCode(vfo VFO, code int) error {
	return c.set("\\set_dcs_code", string(vfo), fmt.Sprintf("%d", code))
}

func (c *RigClient) GetCTCSSSquelch(vfo VFO) (int, error) {
	response, err := c.get("\\get_ctcss_sql", string(vfo))
	if err != nil {
		return 0, err
	}

	sql, err := response.GetInt("CTCSS Sql")
	if err != nil {
		return 0, err
	}

	return sql, nil
}

func (c *RigClient) SetCTCSSSquelch(vfo VFO, sql int) error {
	return c.set("\\set_ctcss_sql", string(vfo), fmt.Sprintf("%d", sql))
}

func (c *RigClient) GetDCSSquelch(vfo VFO) (int, error) {
	response, err := c.get("\\get_dcs_sql", string(vfo))
	if err != nil {
		return 0, err
	}

	sql, err := response.GetInt("DCS Sql")
	if err != nil {
		return 0, err
	}

	return sql, nil
}

func (c *RigClient) SetDCSSquelch(vfo VFO, sql int) error {
	return c.set("\\set_dcs_sql", string(vfo), fmt.Sprintf("%d", sql))
}

// Tuning Step Commands

func (c *RigClient) GetTuningStep(vfo VFO) (int, error) {
	response, err := c.get("\\get_ts", string(vfo))
	if err != nil {
		return 0, err
	}

	ts, err := response.GetInt("Tuning Step")
	if err != nil {
		return 0, err
	}

	return ts, nil
}

func (c *RigClient) SetTuningStep(vfo VFO, ts int) error {
	return c.set("\\set_ts", string(vfo), fmt.Sprintf("%d", ts))
}

// Morse/DTMF Commands

func (c *RigClient) SendMorse(morse string) error {
	return c.set("\\send_morse", morse)
}

func (c *RigClient) StopMorse() error {
	return c.set("\\stop_morse")
}

func (c *RigClient) WaitMorse() error {
	return c.set("\\wait_morse")
}

func (c *RigClient) SendDTMF(digits string) error {
	return c.set("\\send_dtmf", digits)
}

func (c *RigClient) ReceiveDTMF() (string, error) {
	response, err := c.getCustom(parseSingleValue, "\\recv_dtmf")
	if err != nil {
		return "", err
	}

	dtmf, err := response.GetString(singleValueKey)
	if err != nil {
		return "", err
	}

	return dtmf, nil
}

func (c *RigClient) GetDCD() (int, error) {
	response, err := c.get("\\get_dcd")
	if err != nil {
		return 0, err
	}

	dcd, err := response.GetInt("DCD")
	if err != nil {
		return 0, err
	}

	return dcd, nil
}

func (c *RigClient) SendVoiceMemory(msgnum int) error {
	return c.set("\\send_voice_mem", fmt.Sprintf("%d", msgnum))
}

func (c *RigClient) Reset(reset int) error {
	return c.set("\\reset", fmt.Sprintf("%d", reset))
}

func (c *RigClient) SetPowerStatus(status int) error {
	return c.set("\\set_powerstat", fmt.Sprintf("%d", status))
}

func (c *RigClient) GetPowerStatus() (int, error) {
	response, err := c.get("\\get_powerstat")
	if err != nil {
		return 0, err
	}

	status, err := response.GetInt("Power Status")
	if err != nil {
		return 0, err
	}

	return status, nil
}

func (c *RigClient) GetInfo() (string, error) {
	response, err := c.get("\\get_info")
	if err != nil {
		return "", err
	}

	info, err := response.GetString("Info")
	if err != nil {
		return "", err
	}

	return info, nil
}

func (c *RigClient) GetRigInfo() (string, error) {
	response, err := c.getCustom(parseSingleValue, "\\get_rig_info")
	if err != nil {
		return "", err
	}

	info, err := response.GetString(singleValueKey)
	if err != nil {
		return "", err
	}

	return info, nil
}

func (c *RigClient) GetVFOInfo(vfo VFO) (Frequency, string, int, int, int, error) {
	response, err := c.get("\\get_vfo_info", string(vfo))
	if err != nil {
		return 0, "", 0, 0, 0, err
	}

	freq, err := response.GetFloat64("Freq")
	if err != nil {
		return 0, "", 0, 0, 0, err
	}

	mode, err := response.GetString("Mode")
	if err != nil {
		return 0, "", 0, 0, 0, err
	}

	width, err := response.GetInt("Width")
	if err != nil {
		return 0, "", 0, 0, 0, err
	}

	split, err := response.GetInt("Split")
	if err != nil {
		return 0, "", 0, 0, 0, err
	}

	satMode, err := response.GetInt("SatMode")
	if err != nil {
		return 0, "", 0, 0, 0, err
	}

	return Frequency(freq), mode, width, split, satMode, nil
}

func (c *RigClient) DumpState() (string, error) {
	response, err := c.getCustom(parseSingleValue, "\\dump_state")
	if err != nil {
		return "", err
	}

	state, err := response.GetString(singleValueKey)
	if err != nil {
		return "", err
	}

	return state, nil
}

func (c *RigClient) DumpCaps() (string, error) {
	response, err := c.getCustom(parseSingleValue, "\\dump_caps")
	if err != nil {
		return "", err
	}

	caps, err := response.GetString(singleValueKey)
	if err != nil {
		return "", err
	}

	return caps, nil
}

func (c *RigClient) DumpConf() (string, error) {
	response, err := c.getCustom(parseSingleValue, "\\dump_conf")
	if err != nil {
		return "", err
	}

	conf, err := response.GetString(singleValueKey)
	if err != nil {
		return "", err
	}

	return conf, nil
}

func (c *RigClient) Power2mW(power float64, frequency Frequency, mode string) (int, error) {
	response, err := c.get("\\power2mW", fmt.Sprintf("%f", power), frequencyToHL(frequency), mode)
	if err != nil {
		return 0, err
	}

	mw, err := response.GetInt("Power mW")
	if err != nil {
		return 0, err
	}

	return mw, nil
}

func (c *RigClient) MW2Power(powerMW int, frequency Frequency, mode string) (float64, error) {
	response, err := c.get("\\mW2power", fmt.Sprintf("%d", powerMW), frequencyToHL(frequency), mode)
	if err != nil {
		return 0, err
	}

	power, err := response.GetFloat64("Power [0.0..1.0]")
	if err != nil {
		return 0, err
	}

	return power, nil
}

func (c *RigClient) SetClock(timestamp time.Time) error {
	return c.set("\\set_clock", timestamp.Format(timeFormat))
}

func (c *RigClient) GetClock() (time.Time, error) {
	response, err := c.getCustom(parseSingleValue, "\\get_clock")
	if err != nil {
		return time.Time{}, err
	}

	timestamp, err := response.GetString(singleValueKey)
	if err != nil {
		return time.Time{}, err
	}
	if timestamp == "0000-00-00T00:00:00.000+00:00" {
		return time.Time{}, nil
	}

	return time.Parse(timeFormat, timestamp)
}

func (c *RigClient) CheckVFOMode() (bool, error) {
	response, err := c.getCustom(parseChkVFO, "\\chk_vfo")
	if err != nil {
		return false, err
	}

	check, err := response.GetBool("ChkVFO")
	if err != nil {
		return false, err
	}

	return check, nil
}

func parseChkVFO(reader *bufio.Reader) (Response, error) {
	response := Response{Data: make(map[string]string)}
	line, err := reader.ReadString(commandDelimiter)
	if err != nil {
		return Response{}, fmt.Errorf("read response: %w", err)
	}
	line = strings.TrimRight(line, "\r\n")

	sepAt := strings.Index(line, keyValueSeparator)
	if sepAt > 0 {
		key := line[:sepAt]
		value := line[sepAt+len(keyValueSeparator):]
		if key == "ChkVFO" {
			response.Data[key] = value
			return response, nil
		}
	}

	return Response{}, fmt.Errorf("invalid chk_vfo reply: %q", line)
}

func (c *RigClient) SetLockMode(locked bool) error {
	return c.set("\\set_lock_mode", boolToHL(locked))
}

func (c *RigClient) GetLockMode() (bool, error) {
	response, err := c.get("\\get_lock_mode")
	if err != nil {
		return false, err
	}

	locked, err := response.GetBool("Locked")
	if err != nil {
		return false, err
	}

	return locked, nil
}

func (c *RigClient) SendRaw(terminator string, rawCommand []byte) (string, error) {
	response, err := c.get("\\send_raw", terminator, fmt.Sprintf("%02x", rawCommand))
	if err != nil {
		return "", err
	}

	answer, err := response.GetString("Send raw answer")
	if err != nil {
		return "", err
	}

	return answer, nil
}

func (c *RigClient) ClientVersion(version string) error {
	return c.set("\\client_version", version)
}

func (c *RigClient) HamlibVersion() (string, error) {
	response, err := c.getCustom(parseSingleValue, "\\hamlib_version")
	if err != nil {
		return "", err
	}

	version, err := response.GetString(singleValueKey)
	if err != nil {
		return "", err
	}

	return version, nil
}

func (c *RigClient) Test() error {
	return c.set("\\test")
}

func (c *RigClient) SetGPIO(gpio int, value int) error {
	return c.set("\\set_gpio", fmt.Sprintf("%d", gpio), fmt.Sprintf("%d", value))
}

func (c *RigClient) GetGPIO(gpio int) (int, error) {
	response, err := c.get("\\get_gpio", fmt.Sprintf("%d", gpio))
	if err != nil {
		return 0, err
	}

	value, err := response.GetInt("0/1")
	if err != nil {
		return 0, err
	}

	return value, nil
}

func (c *RigClient) SetTransceive(transceive int) error {
	return c.set("\\set_trn", fmt.Sprintf("%d", transceive))
}

func (c *RigClient) GetTransceive() (int, error) {
	response, err := c.get("\\get_trn")
	if err != nil {
		return 0, err
	}

	trn, err := response.GetInt("Transceive")
	if err != nil {
		return 0, err
	}

	return trn, nil
}

func (c *RigClient) SetChannel(channel string) error {
	return c.set("\\set_channel", channel)
}

func (c *RigClient) GetChannel(channel string, readOnly int) (string, error) {
	response, err := c.getCustom(parseSingleValue, "\\get_channel", channel, fmt.Sprintf("%d", readOnly))
	if err != nil {
		return "", err
	}

	value, err := response.GetString(singleValueKey)
	if err != nil {
		return "", err
	}

	return value, nil
}

func (c *RigClient) SendCmd(command string) (string, error) {
	response, err := c.get("\\send_cmd", command)
	if err != nil {
		return "", err
	}

	reply, err := response.GetString("Reply")
	if err != nil {
		return "", err
	}

	return reply, nil
}

func (c *RigClient) SendCmdRx(command string, reply int) (string, error) {
	response, err := c.get("\\send_cmd_rx", command, fmt.Sprintf("%d", reply))
	if err != nil {
		return "", err
	}

	value, err := response.GetString("Reply")
	if err != nil {
		return "", err
	}

	return value, nil
}

func (c *RigClient) StopVoiceMemory() error {
	return c.set("\\stop_voice_mem")
}

func (c *RigClient) SetUplink(uplink int) error {
	return c.set("\\uplink", fmt.Sprintf("%d", uplink))
}

func (c *RigClient) SetTwiddle(timeoutSecs int) error {
	return c.set("\\set_twiddle", fmt.Sprintf("%d", timeoutSecs))
}

func (c *RigClient) GetTwiddle() (int, error) {
	response, err := c.get("\\get_twiddle")
	if err != nil {
		return 0, err
	}

	timeout, err := response.GetInt("Timeout (secs)")
	if err != nil {
		return 0, err
	}

	return timeout, nil
}

func (c *RigClient) SetCache(timeoutMsecs int) error {
	return c.set("\\set_cache", fmt.Sprintf("%d", timeoutMsecs))
}

func (c *RigClient) GetCache() (int, error) {
	response, err := c.get("\\get_cache")
	if err != nil {
		return 0, err
	}

	timeout, err := response.GetInt("Timeout (msecs)")
	if err != nil {
		return 0, err
	}

	return timeout, nil
}

func (c *RigClient) GetVFOList() (string, error) {
	response, err := c.get("\\get_vfo_list")
	if err != nil {
		return "", err
	}

	vfos, err := response.GetString("VFOs")
	if err != nil {
		return "", err
	}

	return vfos, nil
}

func (c *RigClient) GetModes() (string, error) {
	response, err := c.get("\\get_modes")
	if err != nil {
		return "", err
	}

	modes, err := response.GetString("Modes")
	if err != nil {
		return "", err
	}

	return modes, nil
}

func (c *RigClient) Halt() error {
	return c.set("\\halt")
}

func (c *RigClient) Pause(seconds int) error {
	return c.set("\\pause", fmt.Sprintf("%d", seconds))
}

func (c *RigClient) Password(password string) error {
	return c.set("\\password", password)
}

func (c *RigClient) SetSeparator(separator string) error {
	return c.set("\\set_separator", separator)
}

func (c *RigClient) GetSeparator() (string, error) {
	response, err := c.get("\\get_separator")
	if err != nil {
		return "", err
	}

	sep, err := response.GetString("Separator")
	if err != nil {
		return "", err
	}

	return sep, nil
}

func (c *RigClient) GetModeBandwidths(mode string) (string, error) {
	response, err := c.getCustom(parseSingleValue, "\\get_mode_bandwidths", mode)
	if err != nil {
		return "", err
	}

	bw, err := response.GetString(singleValueKey)
	if err != nil {
		return "", err
	}

	return bw, nil
}

func (c *RigClient) SetConf(token string, value string) error {
	return c.set("\\set_conf", token, value)
}

func (c *RigClient) GetConf(token string) (string, error) {
	response, err := c.getCustom(parseSingleValue, "\\get_conf", token)
	if err != nil {
		return "", err
	}

	value, err := response.GetString(singleValueKey)
	if err != nil {
		return "", err
	}

	return value, nil
}
