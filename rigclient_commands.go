package hl

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"
)

// setVFOMode is private because RigClient works exclusively in the VFO mode.
func (c *RigClient) setVFOMode(enabled bool) error {
	return c.set("\\set_vfo_opt", boolToHL(enabled))
}

// GetFrequency retrieves the current frequency of the specified VFO in Hz.
//
// The vfo parameter selects which VFO to query (e.g. VFOA, VFOB, CurrVFO).
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

// SetFrequency sets the frequency of the specified VFO.
//
// The vfo parameter selects which VFO to set (e.g. VFOA, VFOB, CurrVFO).
// The frequency parameter is the desired frequency in Hz.
func (c *RigClient) SetFrequency(vfo VFO, frequency Frequency) error {
	return c.set("\\set_freq", string(vfo), frequencyToHL(frequency))
}

// GetMode retrieves the current operating mode and passband width of the specified VFO.
//
// The vfo parameter selects which VFO to query.
//
// It returns the operating mode (e.g. ModeUSB, ModeLSB, ModeCW, ModeFM, ModeAM)
// and the passband width in Hz.
func (c *RigClient) GetMode(vfo VFO) (Mode, Bandwidth, error) {
	response, err := c.get("\\get_mode", string(vfo))
	if err != nil {
		return "", 0, err
	}

	mode, err := response.GetString("Mode")
	if err != nil {
		return "", 0, err
	}

	passband, err := response.GetFloat64("Passband")
	if err != nil {
		return "", 0, err
	}

	return Mode(mode), Bandwidth(passband), nil
}

// SetMode sets the operating mode and passband width for the specified VFO.
// If lock mode is enabled (see SetLockMode), the command returns successfully without
// actually changing the mode.
//
// The vfo parameter selects which VFO to configure.
// The mode parameter is the desired operating mode (e.g. ModeUSB, ModeLSB, ModeCW, ModeFM, ModeAM).
// The passband parameter is the desired passband width in Hz; use 0 to select the rig's default
// passband for the given mode, or -1 to keep the passband width as it is..
func (c *RigClient) SetMode(vfo VFO, mode Mode, passband Bandwidth) error {
	return c.set("\\set_mode", string(vfo), string(mode), frequencyToHL(passband))
}

// GetVFO retrieves the currently selected VFO.
//
// It returns the active VFO identifier (e.g. VFOA, VFOB, MainVFO, SubVFO).
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

// SetVFO selects the active VFO.
//
// The vfo parameter specifies which VFO to make active (e.g. VFOA, VFOB, MainVFO, SubVFO).
func (c *RigClient) SetVFO(vfo VFO) error {
	return c.set("\\set_vfo", string(vfo))
}

// GetRIT retrieves the current RIT (Receiver Incremental Tuning) offset for the specified VFO.
// RIT allows the receive frequency to be shifted independently from the transmit frequency.
//
// The vfo parameter selects which VFO to query.
//
// It returns the RIT offset in Hz. The value is signed: positive values shift the receive
// frequency up, negative values shift it down.
func (c *RigClient) GetRIT(vfo VFO) (Frequency, error) {
	response, err := c.get("\\get_rit", string(vfo))
	if err != nil {
		return 0, err
	}

	rit, err := response.GetFloat64("RIT")
	if err != nil {
		return 0, err
	}

	return Frequency(rit), nil
}

// SetRIT sets the RIT (Receiver Incremental Tuning) offset for the specified VFO.
//
// The vfo parameter selects which VFO to configure.
// The rit parameter is the desired RIT offset in Hz (signed).
func (c *RigClient) SetRIT(vfo VFO, rit Frequency) error {
	return c.set("\\set_rit", string(vfo), frequencyToHL(rit))
}

// GetXIT retrieves the current XIT (Transmitter Incremental Tuning) offset for the specified VFO.
// XIT allows the transmit frequency to be shifted independently from the receive frequency.
//
// The vfo parameter selects which VFO to query.
//
// It returns the XIT offset in Hz. The value is signed: positive values shift the transmit
// frequency up, negative values shift it down.
func (c *RigClient) GetXIT(vfo VFO) (Frequency, error) {
	response, err := c.get("\\get_xit", string(vfo))
	if err != nil {
		return 0, err
	}

	xit, err := response.GetFloat64("XIT")
	if err != nil {
		return 0, err
	}

	return Frequency(xit), nil
}

// SetXIT sets the XIT (Transmitter Incremental Tuning) offset for the specified VFO.
//
// The vfo parameter selects which VFO to configure.
// The xit parameter is the desired XIT offset in Hz (signed).
func (c *RigClient) SetXIT(vfo VFO, xit Frequency) error {
	return c.set("\\set_xit", string(vfo), frequencyToHL(xit))
}

// GetPTT retrieves the current Push-To-Talk (PTT) status for the specified VFO.
//
// The vfo parameter selects which VFO to query.
//
// It returns the PTT status:
//   - PTTOff: PTT is off (receiving)
//   - PTTOn: PTT is on (transmitting, unspecified source)
//   - PTTOnMic: PTT is on via microphone input
//   - PTTOnData: PTT is on via data input
func (c *RigClient) GetPTT(vfo VFO) (PTTStatus, error) {
	response, err := c.get("\\get_ptt", string(vfo))
	if err != nil {
		return 0, err
	}

	ptt, err := response.GetInt("PTT")
	if err != nil {
		return 0, err
	}

	return PTTStatus(ptt), nil
}

// SetPTT sets the Push-To-Talk (PTT) status for the specified VFO, controlling whether
// the rig is transmitting or receiving.
//
// The vfo parameter selects which VFO to control.
// The ptt parameter specifies the desired PTT state:
//   - PTTOff: stop transmitting
//   - PTTOn: start transmitting (unspecified source)
//   - PTTOnMic: start transmitting via microphone input
//   - PTTOnData: start transmitting via data input
func (c *RigClient) SetPTT(vfo VFO, ptt PTTStatus) error {
	return c.set("\\set_ptt", string(vfo), fmt.Sprintf("%d", ptt))
}

// GetSplitVFO retrieves the split operation status and the designated TX VFO.
// Split operation allows transmitting on a different frequency/VFO than the one used
// for receiving, which is commonly used for working DX pileups.
//
// The vfo parameter selects which VFO to query.
//
// It returns whether split is enabled and the VFO used for transmitting.
func (c *RigClient) GetSplitVFO(vfo VFO) (bool, VFO, error) {
	response, err := c.get("\\get_split_vfo", string(vfo))
	if err != nil {
		return false, "", err
	}

	split, err := response.GetBool("Split")
	if err != nil {
		return false, "", err
	}

	txVFO, err := response.GetString("TX VFO")
	if err != nil {
		return false, "", err
	}

	return split, VFO(txVFO), nil
}

// SetSplitVFO enables or disables split operation and designates the TX VFO.
//
// The vfo parameter selects which VFO context to use.
// The split parameter enables (true) or disables (false) split operation.
// The txVFO parameter specifies which VFO to use for transmitting when split is enabled.
func (c *RigClient) SetSplitVFO(vfo VFO, split bool, txVFO VFO) error {
	return c.set("\\set_split_vfo", string(vfo), boolToHL(split), string(txVFO))
}

// GetSplitFrequency retrieves the split transmit frequency in Hz.
// This is the frequency used for transmitting when split operation is enabled.
//
// The vfo parameter selects which VFO context to query.
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

// SetSplitFrequency sets the split transmit frequency.
//
// The vfo parameter selects which VFO context to use.
// The txFrequency parameter is the desired transmit frequency in Hz.
func (c *RigClient) SetSplitFrequency(vfo VFO, txFrequency Frequency) error {
	return c.set("\\set_split_freq", string(vfo), frequencyToHL(txFrequency))
}

// GetSplitMode retrieves the split transmit mode and passband width.
//
// The vfo parameter selects which VFO context to query.
//
// It returns the transmit mode and passband width in Hz used when split operation is enabled.
func (c *RigClient) GetSplitMode(vfo VFO) (Mode, Bandwidth, error) {
	response, err := c.get("\\get_split_mode", string(vfo))
	if err != nil {
		return "", 0, err
	}

	mode, err := response.GetString("TX Mode")
	if err != nil {
		return "", 0, err
	}

	passband, err := response.GetFloat64("TX Passband")
	if err != nil {
		return "", 0, err
	}

	return Mode(mode), Bandwidth(passband), nil
}

// SetSplitMode sets the split transmit mode and passband width.
//
// The vfo parameter selects which VFO context to use.
// The txMode parameter is the desired transmit mode.
// The txPassband parameter is the desired transmit passband width in Hz; use 0 for the
// rig's default passband for the given mode.
func (c *RigClient) SetSplitMode(vfo VFO, txMode Mode, txPassband Bandwidth) error {
	return c.set("\\set_split_mode", string(vfo), string(txMode), frequencyToHL(txPassband))
}

// GetSplitFreqMode retrieves the split transmit frequency, mode, and passband width in
// a single command. This is more efficient than calling GetSplitFrequency and GetSplitMode
// separately.
//
// The vfo parameter selects which VFO context to query.
//
// It returns the transmit frequency in Hz, the transmit mode, and the transmit passband
// width in Hz.
func (c *RigClient) GetSplitFreqMode(vfo VFO) (Frequency, Mode, Bandwidth, error) {
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

	return Frequency(freq), Mode(mode), Bandwidth(passband), nil
}

// SetSplitFreqMode sets the split transmit frequency, mode, and passband width in a single
// command. This is more efficient than calling SetSplitFrequency and SetSplitMode separately.
//
// The vfo parameter selects which VFO context to use.
// The txFrequency parameter is the desired transmit frequency in Hz.
// The txMode parameter is the desired transmit mode.
// The txPassband parameter is the desired transmit passband width in Hz; use 0 for the
// rig's default passband for the given mode.
func (c *RigClient) SetSplitFreqMode(vfo VFO, txFrequency Frequency, txMode Mode, txPassband Bandwidth) error {
	return c.set("\\set_split_freq_mode", string(vfo), frequencyToHL(txFrequency), string(txMode), frequencyToHL(txPassband))
}

// GetAntenna retrieves antenna information for the specified VFO.
//
// The vfo parameter selects which VFO to query.
// The antCurr parameter specifies which antenna to query. Antenna values are bitmask-based:
// 0 (none), 1 (ANT1), 2 (ANT2), 4 (ANT3), 8 (ANT4), 16 (ANT5).
//
// It returns four values:
//   - ant: the current antenna number
//   - option: an antenna option value (rig-dependent, for example the RX only flag with Icom rigs)
//   - antTx: the antenna selected for transmitting
//   - antRx: the antenna selected for receiving
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

// SetAntenna selects the antenna and sets an antenna option for the specified VFO.
//
// The vfo parameter selects which VFO to configure.
// The antenna parameter is the antenna number to select (bitmask: 1=ANT1, 2=ANT2, 4=ANT3,
// 8=ANT4, 16=ANT5).
// The option parameter is a rig-dependent antenna option value.
func (c *RigClient) SetAntenna(vfo VFO, antenna int, option int) error {
	return c.set("\\set_ant", string(vfo), fmt.Sprintf("%d", antenna), fmt.Sprintf("%d", option))
}

// GetFunc retrieves the on/off status of a rig function for the specified VFO.
// Rig functions are boolean settings that can be toggled on or off.
//
// The vfo parameter selects which VFO to query.
// The function parameter specifies which function to query. Use GetAvailableFunctions
// to discover which functions are supported by the rig. Common functions include:
// FAGC, NB, COMP, VOX, TONE, TSQL, SBKIN, FBKIN, ANF, NR, AIP, APF, MON, MN, RF, ARO,
// LOCK, MUTE, VSC, REV, SQL, ABM, BC, MBC, RIT, AFC, SATMODE, SCOPE, RESUME, TBURST,
// TUNER, XIT, NB2, DSQL, AFLT, ANL, BC2, DUAL_WATCH, DIVERSITY, CSQL, SCEN, SLICE,
// TRANSCEIVE, SPECTRUM, SPECTRUM_HOLD, SEND_MORSE, SEND_VOICE_MEM, OVF_STATUS, SYNC.
func (c *RigClient) GetFunc(vfo VFO, function Function) (bool, error) {
	value, err := c.getSingleValue("\\get_func", string(vfo), string(function))
	if err != nil {
		return false, err
	}

	status, err := parseBool(value)
	if err != nil {
		return false, err
	}

	return status, nil
}

// GetAvailableFunctions queries the rig for the list of functions it supports on the
// specified VFO. The returned list can be used to determine valid values for GetFunc
// and SetFunc.
//
// The vfo parameter selects which VFO to query.
func (c *RigClient) GetAvailableFunctions(vfo VFO) ([]Function, error) {
	functionsString, err := c.getSingleValue("\\get_func", string(vfo), "?")
	if err != nil {
		return nil, err
	}

	parts := strings.Split(strings.TrimSpace(functionsString), " ")
	functions := make([]Function, len(parts))
	for i := range parts {
		functions[i] = Function(parts[i])
	}
	slices.Sort(functions)

	return functions, nil
}

// SetFunc enables or disables a rig function for the specified VFO.
//
// The vfo parameter selects which VFO to configure.
// The function parameter specifies which function to toggle (see GetFunc for the list of
// available function tokens).
// The status parameter enables (true) or disables (false) the function.
func (c *RigClient) SetFunc(vfo VFO, function Function, status bool) error {
	return c.set("\\set_func", string(vfo), string(function), boolToHL(status))
}

// GetLevel retrieves the value of a rig level setting for the specified VFO.
// Levels are analog/graduated settings (as opposed to functions which are boolean).
//
// The vfo parameter selects which VFO to query.
// The level parameter specifies which level to read. Use GetAvailableLevels to discover
// which levels are supported by the rig. Common levels include:
// PREAMP, ATT, AF, RF, SQL, IF, NR, CWPITCH, RFPOWER, MICGAIN, KEYSPD, COMP, AGC,
// VOXGAIN, VOXDELAY, ANTIVOX, PBT_IN, PBT_OUT, NOTCHF, SLOPE_LOW, SLOPE_HIGH,
// BKIN_DLYMS, BAL, METER, RAWSTR, SWR, ALC, STRENGTH, RFPOWER_METER, COMP_METER,
// VD_METER, ID_METER, MONITOR_GAIN, NB, SPECTRUM_MODE, SPECTRUM_SPAN, SPECTRUM_REF,
// SPECTRUM_AVG, SPECTRUM_ATT, TEMP_METER, BAND_SELECT, USB_AF, AGC_TIME, APF.
//
// The returned value is a float64 whose meaning depends on the specific level.
// Many levels use a normalized 0.0 to 1.0 range, while others (like CWPITCH, KEYSPD,
// STRENGTH) use absolute values.
func (c *RigClient) GetLevel(vfo VFO, level Level) (float64, error) {
	valueString, err := c.getSingleValue("\\get_level", string(vfo), string(level))
	if err != nil {
		return 0, err
	}

	value, err := parseFloat(valueString)
	if err != nil {
		return 0, err
	}

	return value, nil
}

// GetAvailableLevels queries the rig for the list of levels it supports on the specified
// VFO. The returned list can be used to determine valid values for GetLevel and SetLevel.
//
// The vfo parameter selects which VFO to query.
func (c *RigClient) GetAvailableLevels(vfo VFO) ([]Level, error) {
	levelsString, err := c.getSingleValue("\\get_level", string(vfo), "?")
	if err != nil {
		return nil, err
	}

	parts := strings.Split(strings.TrimSpace(levelsString), " ")
	levels := make([]Level, len(parts))
	for i := range parts {
		levels[i] = Level(parts[i])
	}
	slices.Sort(levels)

	return levels, nil
}

// SetLevel sets the value of a rig level setting for the specified VFO.
//
// The vfo parameter selects which VFO to configure.
// The level parameter specifies which level to set (see GetLevel for the list of
// available level tokens).
// The value parameter is the desired level value. The meaning depends on the specific
// level; many use a normalized 0.0 to 1.0 range, while others use absolute values.
func (c *RigClient) SetLevel(vfo VFO, level Level, value float64) error {
	return c.set("\\set_level", string(vfo), string(level), fmt.Sprintf("%f", value))
}

// GetParm retrieves the value of a rig-wide parameter. Parameters are global settings
// that are not specific to a VFO, unlike levels and functions.
//
// The parameter specifies which parameter to read. Use GetAvailableParameters to discover
// which parameters are supported by the rig. Common parameters include:
// ANN, APO, BACKLIGHT, BEEP, TIME, BAT, KEYLIGHT, SCREENSAVER, AFIF, BANDSELECT,
// KEYERTYPE, AFIF_LAN, AFIF_WLAN, AFIF_ACC.
func (c *RigClient) GetParm(parameter Parameter) (string, error) {
	return c.getSingleValue("\\get_parm", string(parameter))
}

// GetAvailableParameters queries the rig for the list of rig-wide parameters it supports.
// The returned list can be used to determine valid values for GetParm and SetParm.
func (c *RigClient) GetAvailableParameters() ([]Parameter, error) {
	parametersString, err := c.getSingleValue("\\get_parm", "?")
	if err != nil {
		return nil, err
	}

	parts := strings.Split(strings.TrimSpace(parametersString), " ")
	parameters := make([]Parameter, len(parts))
	for i := range parts {
		parameters[i] = Parameter(parts[i])
	}
	slices.Sort(parameters)

	return parameters, nil
}

// SetParm sets the value of a rig-wide parameter.
//
// The parameter specifies which parameter to set (see GetParm for the list of available
// parameter tokens).
// The value parameter is the desired value as a string.
func (c *RigClient) SetParm(parameter Parameter, value string) error {
	return c.set("\\set_parm", string(parameter), value)
}

// GetMemory retrieves the current memory channel number for the specified VFO.
//
// The vfo parameter selects which VFO to query.
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

// SetMemory sets the current memory channel number for the specified VFO.
//
// The vfo parameter selects which VFO to configure.
// The memory parameter is the desired memory channel number.
func (c *RigClient) SetMemory(vfo VFO, memory int) error {
	return c.set("\\set_mem", string(vfo), fmt.Sprintf("%d", memory))
}

// SetBank selects the memory bank for the specified VFO.
//
// The vfo parameter selects which VFO to configure.
// The bank parameter is the desired memory bank number.
func (c *RigClient) SetBank(vfo VFO, bank int) error {
	return c.set("\\set_bank", string(vfo), fmt.Sprintf("%d", bank))
}

// VFOOp performs a VFO operation on the specified VFO.
//
// The vfo parameter selects which VFO to operate on.
// The op parameter specifies the operation to perform. Valid operations are:
//   - "CPY": copy the current VFO frequency to the other VFO
//   - "XCHG": exchange the frequencies of the two VFOs
//   - "FROM_VFO": transfer VFO data to a memory channel
//   - "TO_VFO": transfer memory channel data to the VFO
//   - "MCL": clear the current memory channel
//   - "UP": increment the frequency by one tuning step
//   - "DOWN": decrement the frequency by one tuning step
//   - "BAND_UP": move to the next higher band
//   - "BAND_DOWN": move to the next lower band
//   - "LEFT": select the left VFO (rig-dependent)
//   - "RIGHT": select the right VFO (rig-dependent)
//   - "TUNE": start the antenna tuner
//   - "TOGGLE": toggle between VFO A and VFO B
func (c *RigClient) VFOOp(vfo VFO, op string) error {
	return c.set("\\vfo_op", string(vfo), op)
}

// Scan starts or stops a scan operation on the specified VFO.
//
// The vfo parameter selects which VFO to scan.
// The scanFct parameter specifies the scan function to perform. Valid scan functions are:
//   - StopScan: stop the current scan
//   - ScanMemory ("MEM"): scan through memory channels
//   - ScanSelected ("SLCT"): scan selected memory channels
//   - ScanPrio ("PRIO"): priority channel scan
//   - ScanProg ("PROG"): programmed scan between frequency limits
//   - ScanDelta ("DELTA"): delta-f scan around the current frequency
//   - ScanVFO ("VFO"): VFO frequency scan
//   - ScanPipeline ("PLT"): priority channel scan (pipeline variant)
//
// The scanChannel parameter is the channel argument for the scan function, whose meaning
// depends on the scan type (e.g. number of channels to scan, or a specific channel number).
func (c *RigClient) Scan(vfo VFO, scanFct ScanFunction, scanChannel int) error {
	return c.set("\\scan", string(vfo), string(scanFct), fmt.Sprintf("%d", scanChannel))
}

// GetRepeaterShift retrieves the current repeater shift direction for the specified VFO.
//
// The vfo parameter selects which VFO to query.
//
// It returns the shift direction as a string: "+" (positive shift), "-" (negative shift),
// or "None" (no shift / simplex).
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

// SetRepeaterShift sets the repeater shift direction for the specified VFO.
//
// The vfo parameter selects which VFO to configure.
// The shift parameter is the desired shift direction: "+" (positive shift),
// "-" (negative shift), or "None" (no shift / simplex).
func (c *RigClient) SetRepeaterShift(vfo VFO, shift string) error {
	return c.set("\\set_rptr_shift", string(vfo), shift)
}

// GetRepeaterOffset retrieves the current repeater offset frequency for the specified VFO.
// The repeater offset is used together with the repeater shift direction to calculate the
// transmit frequency when operating through a repeater.
//
// The vfo parameter selects which VFO to query.
//
// It returns the repeater offset in Hz.
func (c *RigClient) GetRepeaterOffset(vfo VFO) (Frequency, error) {
	response, err := c.get("\\get_rptr_offs", string(vfo))
	if err != nil {
		return 0, err
	}

	offset, err := response.GetFloat64("Rptr Offset")
	if err != nil {
		return 0, err
	}

	return Frequency(offset), nil
}

// SetRepeaterOffset sets the repeater offset frequency for the specified VFO.
//
// The vfo parameter selects which VFO to configure.
// The offset parameter is the desired repeater offset in Hz.
func (c *RigClient) SetRepeaterOffset(vfo VFO, offset Frequency) error {
	return c.set("\\set_rptr_offs", string(vfo), frequencyToHL(offset))
}

// GetCTCSSCode retrieves the current CTCSS (Continuous Tone-Coded Squelch System) sub-audible
// tone used for transmit on the specified VFO.
//
// The vfo parameter selects which VFO to query.
//
// It returns the CTCSS tone in Hz with precision of tenths of Hz, for example 88.5Hz.
func (c *RigClient) GetCTCSSCode(vfo VFO) (Frequency, error) {
	response, err := c.get("\\get_ctcss_tone", string(vfo))
	if err != nil {
		return 0, err
	}

	code, err := response.GetInt("CTCSS Tone")
	if err != nil {
		return 0, err
	}

	return Frequency(float64(code) / 10.0), nil
}

// SetCTCSSCode sets the CTCSS sub-audible tone for transmit on the specified VFO.
//
// The vfo parameter selects which VFO to configure.
// The code parameter is the desired CTCSS tone in Hz with a precision of tenths of Hz, for example 88.5Hz.
func (c *RigClient) SetCTCSSCode(vfo VFO, code Frequency) error {
	return c.set("\\set_ctcss_tone", string(vfo), fmt.Sprintf("%d", int(code/10.0)))
}

// GetDCSCode retrieves the current DCS (Digital-Coded Squelch) code used for transmit
// on the specified VFO.
//
// The vfo parameter selects which VFO to query.
//
// It returns the DCS code as an integer.
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

// SetDCSCode sets the DCS (Digital-Coded Squelch) code for transmit on the specified VFO.
//
// The vfo parameter selects which VFO to configure.
// The code parameter is the desired DCS code.
func (c *RigClient) SetDCSCode(vfo VFO, code int) error {
	return c.set("\\set_dcs_code", string(vfo), fmt.Sprintf("%d", code))
}

// GetCTCSSSquelch retrieves the current CTCSS squelch tone used for receive on the
// specified VFO. This is the tone the receiver listens for to open the squelch.
//
// The vfo parameter selects which VFO to query.
//
// It returns the CTCSS squelch tone in Hz with a precision tenths of Hz.
func (c *RigClient) GetCTCSSSquelch(vfo VFO) (Frequency, error) {
	response, err := c.get("\\get_ctcss_sql", string(vfo))
	if err != nil {
		return 0, err
	}

	sql, err := response.GetInt("CTCSS Sql")
	if err != nil {
		return 0, err
	}

	return Frequency(float64(sql) / 10.0), nil
}

// SetCTCSSSquelch sets the CTCSS squelch tone for receive on the specified VFO.
//
// The vfo parameter selects which VFO to configure.
// The sql parameter is the desired CTCSS squelch tone in Hz with a precision of tenths of Hz.
func (c *RigClient) SetCTCSSSquelch(vfo VFO, sql Frequency) error {
	return c.set("\\set_ctcss_sql", string(vfo), fmt.Sprintf("%d", int(sql*10.0)))
}

// GetDCSSquelch retrieves the current DCS squelch code used for receive on the specified VFO.
// This is the DCS code the receiver listens for to open the squelch.
//
// The vfo parameter selects which VFO to query.
//
// It returns the DCS squelch code as an integer.
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

// SetDCSSquelch sets the DCS squelch code for receive on the specified VFO.
//
// The vfo parameter selects which VFO to configure.
// The sql parameter is the desired DCS squelch code.
func (c *RigClient) SetDCSSquelch(vfo VFO, sql int) error {
	return c.set("\\set_dcs_sql", string(vfo), fmt.Sprintf("%d", sql))
}

// GetTuningStep retrieves the current tuning step size for the specified VFO.
// The tuning step determines how much the frequency changes with each step of the
// VFO dial or UP/DOWN operation.
//
// The vfo parameter selects which VFO to query.
//
// It returns the tuning step size in Hz.
func (c *RigClient) GetTuningStep(vfo VFO) (Frequency, error) {
	response, err := c.get("\\get_ts", string(vfo))
	if err != nil {
		return 0, err
	}

	ts, err := response.GetFloat64("Tuning Step")
	if err != nil {
		return 0, err
	}

	return Frequency(ts), nil
}

// SetTuningStep sets the tuning step size for the specified VFO.
//
// The vfo parameter selects which VFO to configure.
// The ts parameter is the desired tuning step size in Hz.
func (c *RigClient) SetTuningStep(vfo VFO, ts Frequency) error {
	return c.set("\\set_ts", string(vfo), frequencyToHL(ts))
}

// SendMorse sends a Morse code string to the rig for CW transmission. The rig will key
// the transmitter and send the text as Morse code. Spaces in the string are preserved.
//
// The morse parameter is the text to send as Morse code.
func (c *RigClient) SendMorse(morse string) error {
	return c.set("\\send_morse", morse)
}

// StopMorse immediately stops any Morse code transmission in progress.
func (c *RigClient) StopMorse() error {
	return c.set("\\stop_morse")
}

// WaitMorse blocks until the current Morse code transmission completes.
func (c *RigClient) WaitMorse() error {
	return c.set("\\wait_morse")
}

// SendDTMF sends DTMF (Dual-Tone Multi-Frequency) digits through the rig.
//
// The digits parameter is a string of DTMF characters to transmit.
func (c *RigClient) SendDTMF(digits string) error {
	return c.set("\\send_dtmf", digits)
}

// ReceiveDTMF retrieves any received DTMF digits from the rig.
func (c *RigClient) ReceiveDTMF() (string, error) {
	return c.getSingleValue("\\recv_dtmf")
}

// GetDCD retrieves the DCD (Data Carrier Detect) status, which indicates whether
// the squelch is open (a signal is detected on the current frequency).
//
// It returns true if a signal is detected (squelch open), false otherwise.
func (c *RigClient) GetDCD() (bool, error) {
	response, err := c.get("\\get_dcd")
	if err != nil {
		return false, err
	}

	dcd, err := response.GetBool("DCD")
	if err != nil {
		return false, err
	}

	return dcd, nil
}

// SendVoiceMemory plays back a previously recorded voice memory from the rig.
//
// The msgnum parameter is the voice memory slot number to play.
func (c *RigClient) SendVoiceMemory(msgnum int) error {
	return c.set("\\send_voice_mem", fmt.Sprintf("%d", msgnum))
}

// Reset performs a reset operation on the rig.
//
// The mode parameter specifies the type of reset to perform:
//   - ResetNone (0): no reset
//   - ResetSoft (1): software reset
//   - ResetVFO (2): reset VFO settings
//   - ResetMCall (4): reset memory call channel
//   - ResetMaster (8): full master reset (factory defaults)
func (c *RigClient) Reset(mode ResetMode) error {
	return c.set("\\reset", fmt.Sprintf("%d", mode))
}

// SetPowerStatus sets the power status of the rig, allowing you to power the rig on/off
// or change its power state remotely.
//
// The status parameter specifies the desired power state:
//   - PowerOff (0): power off the rig
//   - PowerOn (1): power on the rig
//   - PowerStandby (2): put the rig in standby mode
//   - PowerOperate (4): set the rig to operate mode
func (c *RigClient) SetPowerStatus(status PowerStatus) error {
	return c.set("\\set_powerstat", fmt.Sprintf("%d", status))
}

// GetPowerStatus retrieves the current power status of the rig.
//
// It returns the power state as a PowerStatus value. This command is one of the few
// that can be executed even when the rig is powered off.
func (c *RigClient) GetPowerStatus() (PowerStatus, error) {
	response, err := c.get("\\get_powerstat")
	if err != nil {
		return 0, err
	}

	status, err := response.GetInt("Power Status")
	if err != nil {
		return 0, err
	}

	return PowerStatus(status), nil
}

// GetInfo retrieves miscellaneous information from the rig, such as the firmware version
// or model identification. The format and content of the returned string is rig-dependent.
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

// GetRigInfo retrieves comprehensive rig information as a formatted multi-line string.
// This includes VFO frequencies, modes, bandwidths, split status, and other rig state
// information in a structured format.
func (c *RigClient) GetRigInfo() (string, error) {
	return c.getSingleValue("\\get_rig_info")
}

func parseRigInfo(s string) (RigInfo, error) {
	lines := strings.Split(s, singleValueLineDelimiter)
	if len(lines) == 0 {
		return RigInfo{}, fmt.Errorf("empty input")
	}

	result := RigInfo{}

	for _, line := range lines {
		if value, ok := strings.CutPrefix(line, "VFO="); ok {
			vfoInfo, err := parseRigInfoVFO(value)
			if err != nil {
				return RigInfo{}, err
			}
			result.VFOs = append(result.VFOs, vfoInfo)
		}
		if strings.HasPrefix(line, "Split=") {
			parts := strings.Split(line, " ")
			if len(parts) != 2 {
				return RigInfo{}, fmt.Errorf("get_rig_info: unexpected line: %s", line)
			}
			splitString, ok := strings.CutPrefix(parts[0], "Split=")
			if !ok {
				return RigInfo{}, fmt.Errorf("get_rig_info: unexpected line: %s", line)
			}
			split, err := parseBool(splitString)
			if err != nil {
				return RigInfo{}, fmt.Errorf("get_rig_info: unexpected line: %w", err)
			}
			result.SplitActive = split
			satModeString, ok := strings.CutPrefix(parts[1], "SatMode=")
			if !ok {
				return RigInfo{}, fmt.Errorf("get_rig_info: unexpected line: %s", line)
			}
			satMode, err := parseBool(satModeString)
			if err != nil {
				return RigInfo{}, fmt.Errorf("get_rig_info: unexpected line: %w", err)
			}
			result.SATModeActive = satMode
		}
		if value, ok := strings.CutPrefix(line, "Rig="); ok {
			result.Rig = value
			continue
		}
		if value, ok := strings.CutPrefix(line, "App="); ok {
			result.App = value
			continue
		}
		if value, ok := strings.CutPrefix(line, "Version="); ok {
			result.Version = value
			continue
		}
		if value, ok := strings.CutPrefix(line, "Model="); ok {
			result.Model = value
			continue
		}
		if value, ok := strings.CutPrefix(line, "CRC="); ok {
			result.CRC = value
			continue
		}
		if value, ok := strings.CutPrefix(line, "RPRT "); ok {
			returnCodeInt, err := parseInt(value)
			if err != nil {
				return RigInfo{}, fmt.Errorf("get_rig_info: invalid report: %w", err)
			}
			returnCode := ReturnCode(returnCodeInt)
			if returnCode != RigOk {
				return RigInfo{}, ReturnCodeAsError(returnCode)
			}
			return result, nil
		}
	}

	// normal, err := extractBandwidthFromPart(lines[1], "Normal=")
	// if err != nil {
	// 	return RigInfo{}, err
	// }
	// narrow, err := extractBandwidthFromPart(lines[2], "Narrow=")
	// if err != nil {
	// 	return RigInfo{}, err
	// }
	// wide, err := extractBandwidthFromPart(lines[3], "Wide=")
	// if err != nil {
	// 	return RigInfo{}, err
	// }

	return RigInfo{}, fmt.Errorf("get_rig_info: missing RPRT line")
}

func parseRigInfoVFO(s string) (RigInfoVFO, error) {
	parts := strings.Split(s, " ")
	if len(parts) != 6 {
		return RigInfoVFO{}, fmt.Errorf("unexpected VFO info: %q", s)
	}

	result := RigInfoVFO{
		VFO: VFO(parts[0]),
	}

	valueString, ok := strings.CutPrefix(parts[1], "Freq=")
	if !ok {
		return RigInfoVFO{}, fmt.Errorf("unexpected VFO info: %q", s)
	}
	frequency, err := parseFloat(valueString)
	if err != nil {
		return RigInfoVFO{}, fmt.Errorf("invalid frequency: %q", valueString)
	}
	result.Frequency = Frequency(frequency)

	valueString, ok = strings.CutPrefix(parts[2], "Mode=")
	if !ok {
		return RigInfoVFO{}, fmt.Errorf("unexpected VFO info: %q", s)
	}
	result.Mode = Mode(valueString)

	valueString, ok = strings.CutPrefix(parts[3], "Width=")
	if !ok {
		return RigInfoVFO{}, fmt.Errorf("unexpected VFO info: %q", s)
	}
	passband, err := parseFloat(valueString)
	if err != nil {
		return RigInfoVFO{}, fmt.Errorf("invalid passband: %q", valueString)
	}
	result.Passband = Bandwidth(passband)

	valueString, ok = strings.CutPrefix(parts[4], "RX=")
	if !ok {
		return RigInfoVFO{}, fmt.Errorf("unexpected VFO info: %q", s)
	}
	rxActive, err := parseBool(valueString)
	if err != nil {
		return RigInfoVFO{}, fmt.Errorf("invalid rx flag: %q", valueString)
	}
	result.RXActive = rxActive

	valueString, ok = strings.CutPrefix(parts[5], "TX=")
	if !ok {
		return RigInfoVFO{}, fmt.Errorf("unexpected VFO info: %q", s)
	}
	txActive, err := parseBool(valueString)
	if err != nil {
		return RigInfoVFO{}, fmt.Errorf("invalid tx flag: %q", valueString)
	}
	result.TXActive = txActive

	return result, nil
}

// GetVFOInfo retrieves detailed information for a specific VFO in a single command.
//
// The vfo parameter selects which VFO to query.
//
// It returns:
//   - freq: the VFO frequency in Hz
//   - mode: the operating mode
//   - width: the passband width in Hz
//   - split: whether split operation is enabled
//   - satMode: whether satellite mode is enabled
func (c *RigClient) GetVFOInfo(vfo VFO) (Frequency, Mode, Bandwidth, bool, bool, error) {
	response, err := c.get("\\get_vfo_info", string(vfo))
	if err != nil {
		return 0, "", 0, false, false, err
	}

	freq, err := response.GetFloat64("Freq")
	if err != nil {
		return 0, "", 0, false, false, err
	}

	mode, err := response.GetString("Mode")
	if err != nil {
		return 0, "", 0, false, false, err
	}

	width, err := response.GetFloat64("Width")
	if err != nil {
		return 0, "", 0, false, false, err
	}

	split, err := response.GetBool("Split")
	if err != nil {
		return 0, "", 0, false, false, err
	}

	satMode, err := response.GetBool("SatMode")
	if err != nil {
		return 0, "", 0, false, false, err
	}

	return Frequency(freq), Mode(mode), Bandwidth(width), split, satMode, nil
}

// DumpState retrieves the full rig state as an extensive multi-line diagnostic dump.
// This includes all capabilities, frequency ranges, modes, levels, functions, and
// current settings. Useful for debugging and rig capability discovery.
func (c *RigClient) DumpState() (string, error) {
	return c.getSingleValue("\\dump_state")
}

// DumpCaps retrieves the rig's capability information, listing all supported features,
// commands, frequency ranges, and modes.
func (c *RigClient) DumpCaps() (string, error) {
	return c.getSingleValue("\\dump_caps")
}

// DumpConf retrieves all configuration parameters and their current values from the rig.
func (c *RigClient) DumpConf() (string, error) {
	return c.getSingleValue("\\dump_conf")
}

// Power2mW converts a normalized power value (0.0 to 1.0) to milliwatts for the given
// frequency and mode. This is useful because hamlib uses normalized power values internally,
// but the actual output power in milliwatts depends on the rig, frequency, and mode.
//
// The power parameter is the normalized power level (0.0 = minimum, 1.0 = maximum).
// The frequency parameter is the operating frequency in Hz.
// The mode parameter is the operating mode.
//
// It returns the equivalent power in milliwatts.
func (c *RigClient) Power2mW(power float64, frequency Frequency, mode Mode) (int, error) {
	response, err := c.get("\\power2mW", fmt.Sprintf("%f", power), frequencyToHL(frequency), string(mode))
	if err != nil {
		return 0, err
	}

	mw, err := response.GetInt("Power mW")
	if err != nil {
		return 0, err
	}

	return mw, nil
}

// MW2Power converts a power value in milliwatts to the normalized power range (0.0 to 1.0)
// for the given frequency and mode. This is the inverse of Power2mW.
//
// The powerMW parameter is the power in milliwatts.
// The frequency parameter is the operating frequency in Hz.
// The mode parameter is the operating mode.
//
// It returns the equivalent normalized power value (0.0 = minimum, 1.0 = maximum).
func (c *RigClient) MW2Power(powerMW int, frequency Frequency, mode Mode) (float64, error) {
	response, err := c.get("\\mW2power", fmt.Sprintf("%d", powerMW), frequencyToHL(frequency), string(mode))
	if err != nil {
		return 0, err
	}

	power, err := response.GetFloat64("Power [0.0..1.0]")
	if err != nil {
		return 0, err
	}

	return power, nil
}

// SetClock sets the rig's internal clock to the specified time.
//
// The timestamp parameter is the desired date and time to set.
func (c *RigClient) SetClock(timestamp time.Time) error {
	return c.set("\\set_clock", timestamp.Format(timeFormat))
}

// GetClock retrieves the rig's internal clock time. If the rig does not have a clock
// or the clock is not set, a zero time.Time is returned.
func (c *RigClient) GetClock() (time.Time, error) {
	timestamp, err := c.getSingleValue("\\get_clock")
	if err != nil {
		return time.Time{}, err
	}

	if timestamp == "0000-00-00T00:00:00.000+00:00" {
		return time.Time{}, nil
	}

	return time.Parse(timeFormat, timestamp)
}

// CheckVFOMode checks whether VFO mode is currently active in the rigctld protocol.
// This command is allowed without password authentication.
//
// It returns true if VFO mode is active, false otherwise.
//
// RigClient enforces the VFO mode.
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

// SetLockMode locks or unlocks mode and frequency changes on the rig. When locked,
// commands like SetMode return successfully without actually changing the rig's settings.
//
// The locked parameter enables (true) or disables (false) the lock.
func (c *RigClient) SetLockMode(locked bool) error {
	return c.set("\\set_lock_mode", boolToHL(locked))
}

// GetLockMode retrieves the current lock mode status.
//
// It returns true if the rig is locked (mode/frequency changes are suppressed), false otherwise.
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

// SendRaw sends raw bytes to the rig with a specified terminator character and returns
// the raw response. This provides low-level access to the rig's command protocol.
//
// The terminator parameter is the character used to detect the end of the rig's response.
// The rawCommand parameter is the raw byte sequence to send to the rig (sent as hex-encoded).
//
// It returns the raw response from the rig.
func (c *RigClient) SendRaw(terminator string, rawCommand []byte) ([]byte, error) {
	response, err := c.get("\\send_raw", terminator, bytesToHL(rawCommand))
	if err != nil {
		return nil, err
	}

	bytesString, err := response.GetString("Send raw answer")
	if err != nil {
		return nil, err
	}

	return parseHexBytes(bytesString)
}

func parseHexBytes(s string) ([]byte, error) {
	parts := strings.Split(strings.TrimSpace(s), " ")

	result := make([]byte, 0, len(parts))
	for _, part := range parts {
		hexValue, ok := strings.CutPrefix(part, "0x")
		if !ok {
			return nil, fmt.Errorf("invalid hex number: %s", part)
		}
		value, err := hex.DecodeString(hexValue)
		if err != nil {
			return nil, err
		}
		result = append(result, value...)
	}

	return result, nil
}

// ClientVersion reports the client's version string to the rigctld server. This is stored
// by the server for backward compatibility tracking and may influence protocol behavior.
//
// The version parameter is the client version string to report.
func (c *RigClient) ClientVersion(version string) error {
	return c.set("\\client_version", version)
}

// HamlibVersion retrieves the hamlib/rigctld version string and copyright information
// from the server.
func (c *RigClient) HamlibVersion() (string, error) {
	return c.getSingleValue("\\hamlib_version")
}

// Test sends a test command to verify the connection to the rigctld server is working.
func (c *RigClient) Test() error {
	return c.set("\\test")
}

// SetGPIO sets the state of a CM108/GPIO pin on the rig.
//
// The gpio parameter is the GPIO pin number.
// The value parameter is the desired pin state (false = low, true = high).
func (c *RigClient) SetGPIO(gpio int, value bool) error {
	return c.set("\\set_gpio", fmt.Sprintf("%d", gpio), boolToHL(value))
}

// GetGPIO retrieves the current state of a CM108/GPIO pin on the rig.
//
// The gpio parameter is the GPIO pin number to query.
//
// It returns the pin state (false = low, true = high).
func (c *RigClient) GetGPIO(gpio int) (bool, error) {
	response, err := c.get("\\get_gpio", fmt.Sprintf("%d", gpio))
	if err != nil {
		return false, err
	}

	value, err := response.GetBool("0/1")
	if err != nil {
		return false, err
	}

	return value, nil
}

// SetTransceive enables or disables transceive mode. When transceive mode is active,
// the rig automatically notifies the client of state changes (frequency, mode, etc.)
// without requiring polling.
//
// The mode parameter specifies the transceive mode:
//   - TransceiveOff (0): disable transceive
//   - TransceiveRig (1): enable rig-initiated transceive
//   - TransceivePoll (2): enable poll-based transceive
func (c *RigClient) SetTransceive(mode TransceiveMode) error {
	return c.set("\\set_trn", fmt.Sprintf("%d", mode))
}

// GetTransceive retrieves the current transceive mode setting.
//
// It returns the active transceive mode (TransceiveOff, TransceiveRig, or TransceivePoll).
func (c *RigClient) GetTransceive() (TransceiveMode, error) {
	response, err := c.get("\\get_trn")
	if err != nil {
		return 0, err
	}

	trn, err := response.GetInt("Transceive")
	if err != nil {
		return 0, err
	}

	return TransceiveMode(trn), nil
}

// SetChannel stores channel data to the rig's memory. The channel data format is
// rig-dependent.
//
// The channel parameter is the channel data to store.
func (c *RigClient) SetChannel(channel string) error {
	return c.set("\\set_channel", channel)
}

// GetChannel retrieves channel data from the rig's memory.
//
// The channel parameter identifies which channel to retrieve.
// The readOnly parameter, when true, retrieves the channel data without side effects
// (e.g. without switching the rig to that channel).
func (c *RigClient) GetChannel(channel string, readOnly bool) (string, error) {
	return c.getSingleValue("\\get_channel", channel, boolToHL(readOnly))
}

// SendCmd sends a raw command given as byte slice directly to the rig and returns the reply.
// This bypasses hamlib's command abstraction and sends the command to the rig's native protocol.
// Spaces in the command are preserved.
//
// The command parameter is the raw command as byte slice to send to the rig.
//
// It returns the raw reply from the rig.
func (c *RigClient) SendCmd(command []byte) ([]byte, error) {
	response, err := c.get("\\send_cmd", string(command))
	if err != nil {
		return nil, err
	}

	bytesString, err := response.GetString("Reply")
	if err != nil {
		return nil, err
	}

	return parseHexBytes(bytesString)
}

// SendCmdRx sends a raw command given as byte slice  directly to the rig and returns the reply.
// Unlike SendCmd, this variant takes an additional parameter that describes the terminator byte
// of the reply.
//
// The command parameter is the raw command byte slice to send to the rig.
// The terminator parameter defines the byte value that used by the rig to terminate its reply.
//
// It returns the raw reply from the rig.
func (c *RigClient) SendCmdRx(command []byte, terminator byte) ([]byte, error) {
	response, err := c.get("\\send_cmd_rx", string(command), fmt.Sprintf("%d", terminator))
	if err != nil {
		return nil, err
	}

	bytesString, err := response.GetString("Reply")
	if err != nil {
		return nil, err
	}

	return parseHexBytes(bytesString)
}

// StopVoiceMemory stops any voice memory playback currently in progress.
func (c *RigClient) StopVoiceMemory() error {
	return c.set("\\stop_voice_mem")
}

// SetUplink sets the uplink mode for satellite operation, selecting which VFO is used
// for the uplink (transmit) path.
//
// The uplink parameter specifies the uplink configuration:
//   - UplinkNone (0): no uplink VFO assignment
//   - UplinkSub (1): use the Sub VFO for uplink
//   - UplinkMain (2): use the Main VFO for uplink
func (c *RigClient) SetUplink(uplink Uplink) error {
	return c.set("\\uplink", fmt.Sprintf("%d", uplink))
}

// SetTwiddle sets the twiddle timeout for dial activity detection. When a user physically
// turns the VFO dial on the rig, hamlib detects this activity and can temporarily suppress
// computer-initiated frequency changes for the specified timeout period.
//
// The timeoutSecs parameter is the timeout in seconds after dial activity before computer
// control resumes. Set to 0 to disable twiddle detection.
func (c *RigClient) SetTwiddle(timeoutSecs int) error {
	return c.set("\\set_twiddle", fmt.Sprintf("%d", timeoutSecs))
}

// GetTwiddle retrieves the current twiddle timeout setting.
//
// It returns the timeout in seconds.
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

// SetCache sets the cache timeout for rig data. Hamlib caches rig state to avoid
// excessive serial communication. This controls how long cached values are considered valid.
//
// The timeoutMsecs parameter is the cache timeout in milliseconds.
func (c *RigClient) SetCache(timeoutMsecs int) error {
	return c.set("\\set_cache", fmt.Sprintf("%d", timeoutMsecs))
}

// GetCache retrieves the current cache timeout setting.
//
// It returns the cache timeout in milliseconds.
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

// GetVFOList retrieves the list of VFOs available on the rig.
//
// It returns a space-separated string of VFO name tokens (e.g. "VFOA VFOB").
func (c *RigClient) GetVFOList() ([]VFO, error) {
	response, err := c.get("\\get_vfo_list")
	if err != nil {
		return nil, err
	}

	reply, err := response.GetString("VFOs")
	if err != nil {
		return nil, err
	}

	// There seems to be a bug in how the VFO list is prepared in rigctld that appends stray bytes to the end of the list.
	reply = strings.TrimRight(reply, " \x01")
	reply = strings.TrimSpace(reply)

	parts := strings.Split(reply, " ")
	vfos := make([]VFO, len(parts))
	for i := range parts {
		vfos[i] = VFO(parts[i])
	}
	slices.Sort(vfos)

	return vfos, nil
}

// GetModes retrieves all operating modes supported by the rig along with their
// normal, narrow, and wide bandwidth values.
//
// It returns a map from Mode to ModeBandwidths, where each ModeBandwidths contains
// the normal, narrow, and wide passband widths in Hz for that mode.
func (c *RigClient) GetModes() (map[Mode]ModeBandwidths, error) {
	modes, err := c.getSingleValue("\\get_modes")
	if err != nil {
		return nil, err
	}
	return parseModes(modes)
}

func parseModes(s string) (map[Mode]ModeBandwidths, error) {
	result := make(map[Mode]ModeBandwidths)
	lines := strings.Split(s, singleValueLineDelimiter)
	if len(lines) == 0 {
		return nil, fmt.Errorf("empty input")
	}

	i := 1
	for i < len(lines) {
		if strings.TrimSpace(lines[i]) == "Bandwidths:" {
			i++
			break
		}
		i++
	}

	for ; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		parts := strings.Split(line, "\t")
		if len(parts) != 4 {
			return nil, fmt.Errorf("invalid mode bandwidth line: %q", line)
		}

		mode := Mode(strings.ToUpper(parts[0]))

		normal, err := extractBandwidthFromPart(parts[1], "Normal: ")
		if err != nil {
			return nil, err
		}
		narrow, err := extractBandwidthFromPart(parts[2], "Narrow: ")
		if err != nil {
			return nil, err
		}
		wide, err := extractBandwidthFromPart(parts[3], "Wide: ")
		if err != nil {
			return nil, err
		}

		result[mode] = ModeBandwidths{
			Normal: normal,
			Narrow: narrow,
			Wide:   wide,
		}
	}

	return result, nil
}

func extractBandwidthFromPart(part string, prefix string) (Bandwidth, error) {
	remainder, ok := strings.CutPrefix(part, prefix)
	if !ok {
		return 0, fmt.Errorf("invalid bandwidth prefix: expected %q but got %q", prefix, part)
	}

	valueString, _ := strings.CutSuffix(remainder, ",")

	return parseBandwidth(valueString)
}

func parseBandwidth(s string) (Bandwidth, error) {
	s = strings.TrimSpace(strings.ToLower(s))

	var valueString string
	var mhz, khz, hz bool
	valueString, mhz = strings.CutSuffix(s, "mhz")
	if !mhz {
		valueString, khz = strings.CutSuffix(s, "khz")
	}
	if !mhz && !khz {
		valueString, hz = strings.CutSuffix(s, "hz")
	}

	valueString = strings.TrimSpace(valueString)
	value, err := strconv.ParseFloat(valueString, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid bandwidth number: %w", err)
	}

	switch {
	case hz:
		return Bandwidth(value), nil
	case khz:
		return Bandwidth(value * 1_000), nil
	case mhz:
		return Bandwidth(value * 1_000_000), nil
	default:
		return 0, fmt.Errorf("unknown bandwidth unit: %q", s)
	}
}

// Halt terminates the rigctld daemon. After this command, the connection will be closed
// and no further commands can be sent.
func (c *RigClient) Halt() error {
	return c.set("\\halt")
}

// Pause instructs the rigctld server to pause execution for the specified number of seconds.
//
// The seconds parameter is the number of seconds to pause.
func (c *RigClient) Pause(seconds int) error {
	return c.set("\\pause", fmt.Sprintf("%d", seconds))
}

// Password authenticates the client with the rigctld server. When rigctld is started
// with password protection, this command must be called before most other commands
// are allowed. Only CheckVFOMode, SetVFO, and Password itself are permitted without
// prior authentication.
//
// The password parameter is the authentication password.
func (c *RigClient) Password(password string) error {
	return c.set("\\password", password)
}

// SetSeparator sets the response field separator character used by the rigctld protocol.
// This changes the character used to separate key-value pairs in responses.
//
// The separator parameter is a single character to use as the field separator.
func (c *RigClient) SetSeparator(separator string) error {
	return c.set("\\set_separator", separator)
}

// GetSeparator retrieves the current response field separator character.
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

// GetModeBandwidths retrieves the normal, narrow, and wide bandwidth values for a
// specific operating mode.
//
// The mode parameter is the mode to query (e.g. ModeUSB, ModeCW, ModeFM).
//
// It returns a ModeBandwidths struct containing the normal, narrow, and wide passband
// widths in Hz for the requested mode.
func (c *RigClient) GetModeBandwidths(mode Mode) (ModeBandwidths, error) {
	bw, err := c.getSingleValue("\\get_mode_bandwidths", string(mode))
	if err != nil {
		return ModeBandwidths{}, err
	}

	return parseModeBandwidths(bw)
}

func parseModeBandwidths(s string) (ModeBandwidths, error) {
	lines := strings.Split(s, singleValueLineDelimiter)
	if len(lines) != 4 {
		return ModeBandwidths{}, fmt.Errorf("empty input")
	}

	mode, ok := strings.CutPrefix(lines[0], "Mode=")
	if !ok {
		return ModeBandwidths{}, fmt.Errorf("invalid mode bandwidths: %q", s)
	}
	normal, err := extractBandwidthFromPart(lines[1], "Normal=")
	if err != nil {
		return ModeBandwidths{}, err
	}
	narrow, err := extractBandwidthFromPart(lines[2], "Narrow=")
	if err != nil {
		return ModeBandwidths{}, err
	}
	wide, err := extractBandwidthFromPart(lines[3], "Wide=")
	if err != nil {
		return ModeBandwidths{}, err
	}

	return ModeBandwidths{
		Mode:   Mode(mode),
		Normal: normal,
		Narrow: narrow,
		Wide:   wide,
	}, nil
}

// SetConf sets a configuration token value on the rig. Configuration tokens are
// rig-specific settings that can be used to fine-tune rig behavior. Use DumpConf
// to discover available configuration tokens.
//
// The token parameter is the configuration token name or numeric ID.
// The value parameter is the desired value as a string.
func (c *RigClient) SetConf(token string, value string) error {
	return c.set("\\set_conf", token, value)
}

// GetConf retrieves the value of a configuration token from the rig.
//
// The token parameter is the configuration token name or numeric ID to query.
//
// It returns the current value as a string.
func (c *RigClient) GetConf(token string) (string, error) {
	keyValue, err := c.getSingleValue("\\get_conf", token)
	if err != nil {
		return "", err
	}

	_, after, found := strings.Cut(keyValue, "=")
	if found {
		return after, nil
	}

	return keyValue, nil
}
