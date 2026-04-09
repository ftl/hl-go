package hl

import (
	"fmt"
	"slices"
	"strconv"

	"github.com/ftl/hamradio"
)

const timeFormat = "2006-01-02T15:04:05.999Z07:00"

// Request to a hamlib daemon.
type Request struct {
	Command string
	Args    []string
}

// Response from a hamlib daemon.
type Response struct {
	CommandEcho string
	Data        map[string]string
	ReturnCode  ReturnCode
}

// GetString returns the data value with the given key as string.
func (r Response) GetString(key string) (string, error) {
	result, ok := r.Data[key]
	if !ok {
		return "", fmt.Errorf("no %s value", key)
	}
	return result, nil
}

// GetInt returns the data value with the given key as int.
func (r Response) GetInt(key string) (int, error) {
	value, ok := r.Data[key]
	if !ok {
		return 0, fmt.Errorf("no %s value", key)
	}
	result, err := parseInt(value)
	if err != nil {
		return 0, fmt.Errorf("cannot parse %s value: %w", key, err)
	}
	return result, nil
}

func parseInt(s string) (int, error) {
	return strconv.Atoi(s)
}

// GetFloat64 returns the data value with the given key as float64.
func (r Response) GetFloat64(key string) (float64, error) {
	value, ok := r.Data[key]
	if !ok {
		return 0, fmt.Errorf("no %s value", key)
	}
	result, err := parseFloat(value)
	if err != nil {
		return 0, fmt.Errorf("cannot parse %s value: %w", key, err)
	}
	return result, nil
}

func parseFloat(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

// GetBool returns the data value with the given key as bool.
func (r Response) GetBool(key string) (bool, error) {
	value, ok := r.Data[key]
	if !ok {
		return false, fmt.Errorf("no %s value", key)
	}
	result, err := parseBool(value)
	if err != nil {
		return false, fmt.Errorf("cannot parse %s value: %w", key, err)
	}
	return result, nil
}

func parseBool(s string) (bool, error) {
	return strconv.ParseBool(s)
}

// ReturnCode represents the numeric error codes returned by the rigctld protocol.
// A value of 0 (RigOk) indicates success; all other values are negative integers
// indicating specific error conditions. These codes map directly to hamlib's
// RIG_OK / RIG_E* error codes.
type ReturnCode int

const (
	// RigOk indicates the command completed successfully (RIG_OK, 0).
	RigOk ReturnCode = iota * -1
	// RigInvalidParameter indicates an invalid parameter was passed to the command (RIG_EINVAL, -1).
	// This typically means a value was syntactically wrong or out of an expected range.
	RigInvalidParameter
	// RigInvalidConfiguration indicates the rig configuration is invalid or incomplete (RIG_ECONF, -2).
	// For example, a required serial port or network address was not configured.
	RigInvalidConfiguration
	// RigMemoryShortage indicates hamlib ran out of memory during the operation (RIG_ENOMEM, -3).
	RigMemoryShortage
	// RigFeatureNotImplemented indicates the requested feature is not implemented in the
	// backend for this rig model (RIG_ENIMPL, -4). The rig hardware may support it, but
	// the hamlib driver does not provide it yet.
	RigFeatureNotImplemented
	// RigCommunicationTimeout indicates the rig did not respond within the expected time
	// (RIG_ETIMEOUT, -5). This usually points to a serial/network connection problem or
	// the rig being powered off.
	RigCommunicationTimeout
	// RigIOError indicates a low-level I/O error occurred while communicating with the rig
	// (RIG_EIO, -6), such as a read/write failure on the serial port or network socket.
	RigIOError
	// RigInternalError indicates an unexpected internal error within hamlib (RIG_EINTERNAL, -7).
	RigInternalError
	// RigProtocolError indicates the response from the rig did not match the expected
	// protocol format (RIG_EPROTO, -8). The data received was malformed or unexpected.
	RigProtocolError
	// RigCommandRejected indicates the rig explicitly rejected the command (RIG_ERJCTED, -9).
	// The rig understood the command but refused to execute it, for example because the rig
	// is in a state that does not allow the operation.
	RigCommandRejected
	// RigArgumentTruncated indicates one or more arguments were truncated before being sent
	// to the rig (RIG_ETRUNC, -10), typically because they exceeded the rig's maximum length.
	RigArgumentTruncated
	// RigFeatureNotAvailable indicates the feature is known to hamlib but is not available
	// on this specific rig model or in the current rig state (RIG_ENAVAIL, -11).
	RigFeatureNotAvailable
	// RigVFONotAccessible indicates the targeted VFO cannot be reached (RIG_ENTARGET, -12).
	// The VFO may not exist on this rig or may not be selectable in the current configuration.
	RigVFONotAccessible
	// RigCommunicationBusError indicates a bus-level error on the communication link
	// (RIG_BUSERROR, -13), such as a CI-V bus error on Icom rigs.
	RigCommunicationBusError
	// RigCommunicationBusCollision indicates a bus collision occurred (RIG_BUSBUSY, -14),
	// meaning another device on the shared bus (e.g. CI-V) transmitted at the same time.
	RigCommunicationBusCollision
	// RigNullHandle indicates a NULL rig handle was passed to a hamlib function (RIG_EARG, -15).
	// This is an internal programming error.
	RigNullHandle
	// RigInvalidVFO indicates the specified VFO is not valid for this rig (RIG_EVFO, -16).
	RigInvalidVFO
	// RigArgumentOutOfDomain indicates an argument is out of its valid domain (RIG_EDOM, -17),
	// for example a frequency outside the rig's supported range.
	RigArgumentOutOfDomain
	// RigFunctionDeprecated indicates the called function has been deprecated and should
	// no longer be used (RIG_EDEPRECATED, -18).
	RigFunctionDeprecated
	// RigSecurityError indicates the command was blocked due to insufficient authentication
	// (RIG_ESECURITY, -19). Call Password first when rigctld requires authentication.
	RigSecurityError
	// RigNotPoweredOn indicates the command failed because the rig is not powered on
	// (RIG_EPOWER, -20). Only a few commands (e.g. GetPowerStatus, SetPowerStatus) work
	// when the rig is off.
	RigNotPoweredOn
	// RigLimitExceeded indicates a resource or rate limit was exceeded (RIG_ELIMIT, -21).
	RigLimitExceeded
	// RigAccessDenied indicates access to the requested resource was denied (RIG_EACCESS, -22).
	RigAccessDenied
)

// VFO identifies a Virtual Frequency Oscillator on the rig. Most rigs have at least two
// VFOs (VFOA and VFOB), allowing the operator to have two frequencies ready. Some rigs
// provide additional VFOs or a Main/Sub VFO arrangement for dual-receiver operation.
//
// Use CurrVFO to target whichever VFO is currently selected on the rig. Use TXVFO and
// RXVFO to target the VFO currently assigned for transmitting or receiving, respectively.
// The Main/Sub and MainA/MainB/SubA/SubB variants are used on rigs with dual-receiver
// architectures (e.g. Icom IC-7610, Yaesu FTDX101).
type VFO string

const (
	// CurrVFO targets whichever VFO is currently selected on the rig. This is the most
	// commonly used VFO value and lets the rig decide which VFO is active.
	CurrVFO VFO = "currVFO"
	// MainAVFO targets VFO A on the main receiver. Used on dual-receiver rigs that have
	// A/B VFOs on each receiver (e.g. Yaesu FTDX101D).
	MainAVFO VFO = "MainA"
	// MainBVFO targets VFO B on the main receiver. Used on dual-receiver rigs that have
	// A/B VFOs on each receiver.
	MainBVFO VFO = "MainB"
	// MainVFO targets the main receiver's VFO. Used on rigs with a Main/Sub receiver
	// architecture where each receiver has a single VFO (e.g. Icom IC-7610).
	MainVFO VFO = "Main"
	// MEM targets the rig's memory instead of a real VFO.
	MEM VFO = "MEM"
	// RXVFO targets whichever VFO is currently used for receiving. When split operation
	// is active, this differs from TXVFO.
	RXVFO VFO = "RX"
	// SubAVFO targets VFO A on the sub receiver. Used on dual-receiver rigs that have
	// A/B VFOs on each receiver.
	SubAVFO VFO = "SubA"
	// SubBVFO targets VFO B on the sub receiver. Used on dual-receiver rigs that have
	// A/B VFOs on each receiver.
	SubBVFO VFO = "SubB"
	// SubVFO targets the sub receiver's VFO. Used on rigs with a Main/Sub receiver
	// architecture.
	SubVFO VFO = "Sub"
	// TXVFO targets whichever VFO is currently used for transmitting. When split operation
	// is active, this differs from RXVFO.
	TXVFO VFO = "TX"
	// VFOA targets VFO A, the primary VFO on most rigs. On simple rigs without Main/Sub
	// receivers, VFOA and VFOB are the two available VFOs.
	VFOA VFO = "VFOA"
	// VFOB targets VFO B, the secondary VFO on most rigs. Commonly used as the transmit
	// VFO during split operation.
	VFOB VFO = "VFOB"
	// VFOC targets VFO C, a third VFO available on some rigs (e.g. certain Icom models).
	VFOC VFO = "VFOC"
)

// Frequency represents a radio frequency in Hz. It is an alias for hamradio.Frequency.
type Frequency = hamradio.Frequency

// Bandwidth represents a passband width in Hz. It is an alias for Frequency, since both
// are measured in the same unit. A bandwidth of 0 typically means "use the rig's default
// passband for the current mode".
type Bandwidth = Frequency

// Mode represents a radio operating mode (modulation type). The mode determines how
// the radio signal is modulated and demodulated. Common modes include USB and LSB for
// single sideband voice, CW for Morse code, FM for frequency modulation, and AM for
// amplitude modulation. The PKT* variants (PKTUSB, PKTLSB, PKTFM) are digital/data
// modes that use the corresponding modulation with the rig's data port. The SAM/SAL/SAH
// variants are synchronous AM modes. Not all rigs support all modes; use
// [RigClient.GetModes] to discover which modes the connected rig supports.
type Mode string

const (
	// ModeAM is Amplitude Modulation. The carrier and both sidebands are transmitted.
	// Primarily used on the AM broadcast bands and some HF frequencies.
	ModeAM Mode = "AM"
	// ModeAMS is Amplitude Modulation Synchronous. A variant of AM that uses a
	// phase-locked carrier for improved reception in fading conditions.
	ModeAMS Mode = "AMS"
	// ModeCW is Continuous Wave, i.e. Morse code using on/off keying of a carrier.
	// The standard mode for Morse code operation.
	ModeCW Mode = "CW"
	// ModeCWR is CW Reverse. CW with the sideband reversed (LSB instead of USB).
	// Useful for separating CW signals or for personal preference.
	ModeCWR Mode = "CWR"
	// ModeDSB is Double Sideband, suppressed carrier. Both sidebands are transmitted
	// without a carrier. Less common than SSB, but used in some specialty applications.
	ModeDSB Mode = "DSB"
	// ModeECSSLSB is Exalted Carrier Single Sideband, Lower Sideband. A technique for
	// receiving AM broadcasts on SSB rigs by reinjecting a carrier on the lower sideband.
	ModeECSSLSB Mode = "ECSSLSB"
	// ModeECSSUSB is Exalted Carrier Single Sideband, Upper Sideband. A technique for
	// receiving AM broadcasts on SSB rigs by reinjecting a carrier on the upper sideband.
	ModeECSSUSB Mode = "ECSSUSB"
	// ModeFA is FAX mode for receiving weather fax and radiofax transmissions.
	ModeFA Mode = "FA"
	// ModeFM is Frequency Modulation. The standard mode for VHF/UHF voice communication
	// and local repeater operation.
	ModeFM Mode = "FM"
	// ModeLSB is Lower Single Sideband. The standard SSB mode for HF frequencies below
	// 10 MHz (e.g. 160m, 80m, 40m). Only the lower sideband is transmitted.
	ModeLSB Mode = "LSB"
	// ModePKTFM is Packet/Digital over FM. Uses the rig's FM modulation with the data
	// port (audio input/output) for digital modes such as packet radio or APRS.
	ModePKTFM Mode = "PKTFM"
	// ModePKTLSB is Packet/Digital over LSB. Uses the rig's LSB modulation with the
	// data port for digital modes such as PSK31, FT8, RTTY, or SSTV.
	ModePKTLSB Mode = "PKTLSB"
	// ModePKTUSB is Packet/Digital over USB. Uses the rig's USB modulation with the
	// data port for digital modes such as PSK31, FT8, RTTY, or SSTV. The most common
	// mode for HF digital operations.
	ModePKTUSB Mode = "PKTUSB"
	// ModeRTTY is Radioteletype (FSK). The rig generates an FSK signal directly, as
	// opposed to PKTUSB/PKTLSB which use audio-based AFSK.
	ModeRTTY Mode = "RTTY"
	// ModeRTTYR is RTTY Reverse. RTTY with the mark and space tones swapped.
	ModeRTTYR Mode = "RTTYR"
	// ModeSAH is Synchronous AM, High sideband. A synchronous AM variant that locks onto
	// the upper sideband of an AM signal to reduce fading and interference.
	ModeSAH Mode = "SAH"
	// ModeSAL is Synchronous AM, Low sideband. A synchronous AM variant that locks onto
	// the lower sideband of an AM signal to reduce fading and interference.
	ModeSAL Mode = "SAL"
	// ModeSAM is Synchronous AM. The rig phase-locks to the AM carrier for improved
	// reception quality, reducing the effects of selective fading.
	ModeSAM Mode = "SAM"
	// ModeUSB is Upper Single Sideband. The standard SSB mode for HF frequencies above
	// 10 MHz (e.g. 20m, 17m, 15m, 10m) and for all VHF/UHF SSB operation. Only the
	// upper sideband is transmitted.
	ModeUSB Mode = "USB"
	// ModeWFM is Wide FM. Used for receiving wideband FM broadcasts (commercial radio
	// stations) which use a wider deviation than narrowband FM.
	ModeWFM Mode = "WFM"
)

// ModeBandwidths holds the normal, narrow, and wide passband widths for a given operating
// mode. These three bandwidth presets are defined by the rig and represent the typical,
// narrowest, and widest filter settings available for the mode.
type ModeBandwidths struct {
	Mode   Mode
	Normal Bandwidth
	Narrow Bandwidth
	Wide   Bandwidth
}

func Modes(m map[Mode]ModeBandwidths) []Mode {
	result := make([]Mode, 0, len(m))
	for mode := range m {
		result = append(result, mode)
	}
	slices.Sort(result)
	return result
}

// Function represents a boolean rig function that can be toggled on or off.
// Functions are discrete on/off settings, as opposed to levels which have graduated
// values. Examples include noise blanker (NB), noise reduction (NR), VOX, tone squelch
// (TSQL), and monitor (MON). Not all rigs support all functions; use
// [RigClient.GetAvailableFunctions] to discover which functions the connected rig supports.
type Function string

const (
	// AdvancedIPFunction enables the Advanced Intercept Point (AIP) function. On Yaesu rigs
	// this improves dynamic range and reduces intermodulation from strong adjacent signals.
	AdvancedIPFunction Function = "AIP"
	// AudioFilterFunction enables an audio filter (AFLT). On some rigs this activates a
	// dedicated DSP audio filter for improved audio quality.
	AudioFilterFunction Function = "AFLT"
	// AudioPeakFilterFunction enables the Audio Peak Filter (APF). Narrows the audio
	// passband to a peak centered on the CW sidetone frequency, improving CW reception
	// in crowded band conditions.
	AudioPeakFilterFunction Function = "APF"
	// AutoBandModeFunction enables Auto Band Mode (ABM). Automatically selects an
	// appropriate mode when the VFO frequency is changed to a new band.
	AutoBandModeFunction Function = "ABM"
	// AutoFrequencyControlFunction enables Automatic Frequency Control (AFC). Continuously
	// corrects the receive frequency to track a drifting signal, commonly used in FM mode.
	AutoFrequencyControlFunction Function = "AFC"
	// AutoNoiseLimiterFunction enables the Automatic Noise Limiter (ANL). Clips impulse
	// noise (ignition interference, static crashes) to reduce its impact on the audio.
	AutoNoiseLimiterFunction Function = "ANL"
	// AutoNotchFilterFunction enables the Automatic Notch Filter (ANF). Automatically
	// identifies and notches out heterodyne tones and carriers in the passband.
	AutoNotchFilterFunction Function = "ANF"
	// AutoRepeaterOffsetFunction enables Automatic Repeater Offset (ARO). Automatically
	// sets the repeater offset based on the operating frequency and regional band plan.
	AutoRepeaterOffsetFunction Function = "ARO"
	// BeatCanceler2Function enables the second Beat Canceller (BC2). A second-generation
	// DSP circuit that removes heterodyne beat tones from the audio.
	BeatCanceler2Function Function = "BC2"
	// BeatCancellerFunction enables the Beat Canceller (BC). A DSP circuit that detects
	// and removes heterodyne beat tones (whistles) from the received audio.
	BeatCancellerFunction Function = "BC"
	// CombinedSquelchFunction enables Combined Squelch (CSQL). A squelch mode that
	// combines CTCSS and DCS, requiring both codes to open the squelch.
	CombinedSquelchFunction Function = "CSQL"
	// CompressorFunction enables the speech compressor. Increases the average power of
	// transmitted voice audio to improve readability at the receiving end.
	CompressorFunction Function = "COMP"
	// DigitalSquelchFunction enables Digital Squelch (DSQL). A squelch mode based on
	// digital code identification, used in digital voice modes such as D-STAR or C4FM.
	DigitalSquelchFunction Function = "DSQL"
	// DiversityFunction enables antenna diversity reception. The rig switches between or
	// combines signals from multiple antennas to reduce fading.
	DiversityFunction Function = "DIVERSITY"
	// DualWatchFunction enables dual watch (sub-receiver monitoring). The rig simultaneously
	// monitors a second frequency on the sub-receiver while operating on the main VFO.
	DualWatchFunction Function = "DUAL_WATCH"
	// FastAGCFunction enables Fast AGC (Automatic Gain Control). Increases the attack speed
	// of the AGC circuit, useful for fast CW or rapid signal changes.
	FastAGCFunction Function = "FAGC"
	// FullBreakInFunction enables Full QSK (full break-in) for CW operation. The rig
	// switches between transmit and receive between every dit and dah, allowing the operator
	// to hear the band between characters.
	FullBreakInFunction Function = "FBKIN"
	// LockFunction locks the VFO dial to prevent accidental frequency changes. The frequency
	// cannot be changed while the lock is active.
	LockFunction Function = "LOCK"
	// ManualBeatCancellerFunction enables the Manual Beat Canceller (MBC). Similar to the
	// automatic beat canceller but requires manual tuning of the notch frequency.
	ManualBeatCancellerFunction Function = "MBC"
	// ManualNotchFunction enables the manual notch filter (MN). A narrow DSP notch whose
	// center frequency is adjusted manually via a notch control.
	ManualNotchFunction Function = "MN"
	// MonitorFunction enables the monitor function (MON). Feeds the transmitted audio back
	// into the receiver's audio path so the operator can monitor the transmitted signal.
	MonitorFunction Function = "MON"
	// MuteFunction mutes the receiver audio output.
	MuteFunction Function = "MUTE"
	// NoiseBlanker2Function enables the second noise blanker circuit (NB2). An additional
	// or alternative noise blanker with different characteristics from the primary NB.
	NoiseBlanker2Function Function = "NB2"
	// NoiseBlankerFunction enables the Noise Blanker (NB). Suppresses impulse noise such
	// as ignition interference by blanking the IF signal during noise pulses.
	NoiseBlankerFunction Function = "NB"
	// NoiseReductionFunction enables DSP Noise Reduction (NR). Applies digital signal
	// processing to reduce background noise and improve signal intelligibility.
	NoiseReductionFunction Function = "NR"
	// OverflowStatusFunction reports the ADC overflow (OVF) status. When active, it
	// indicates that the receiver's ADC is being overloaded by a strong signal.
	OverflowStatusFunction Function = "OVF_STATUS"
	// ResumeFunction controls scan resume behavior. When enabled, the scanner automatically
	// resumes scanning after a squelch opening has closed.
	ResumeFunction Function = "RESUME"
	// ReverseFunction enables the reverse function (REV). On FM/repeater operation, swaps
	// the transmit and receive frequencies to listen on the repeater's input frequency.
	ReverseFunction Function = "REV"
	// RFFunction enables the RF (radio frequency) amplifier or attenuator function.
	// Behavior is rig-dependent; on some rigs this enables the RF preamp.
	RFFunction Function = "RF"
	// RITFunction enables RIT (Receiver Incremental Tuning). Activates the RIT circuit,
	// allowing the receive frequency to be offset from the transmit frequency.
	RITFunction Function = "RIT"
	// SATModeFunction enables satellite mode (SATMODE). Configures the rig for satellite
	// operation, linking the main and sub VFOs so that tuning one automatically compensates
	// the other for Doppler shift.
	SATModeFunction Function = "SATMODE"
	// SceneFunction enables scene memory recall (SCEN). Restores a previously saved
	// set of rig settings (a "scene") in one operation.
	SceneFunction Function = "SCEN"
	// ScopeFunction enables the bandscope/spectrum scope display. Activates the rig's
	// built-in panadapter or spectrum display.
	ScopeFunction Function = "SCOPE"
	// SemiBreakInFunction enables semi-automatic QSK (semi break-in) for CW. The rig
	// switches to receive after a configurable delay following the last transmitted character,
	// rather than immediately (full break-in).
	SemiBreakInFunction Function = "SBKIN"
	// SendMorseFunction enables or disables the rig's internal CW keyer/Morse sending
	// capability, as opposed to using an external keyer.
	SendMorseFunction Function = "SEND_MORSE"
	// SendVoiceMemoryFunction enables or triggers playback of a recorded voice memory
	// message.
	SendVoiceMemoryFunction Function = "SEND_VOICE_MEM"
	// SliceFunction enables receiver slice operation. On SDR-based rigs, a slice is an
	// independent receiver channel within the digitized bandwidth.
	SliceFunction Function = "SLICE"
	// SpectrumFunction enables the spectrum display output. Activates streaming of spectrum
	// data from the rig.
	SpectrumFunction Function = "SPECTRUM"
	// SpectrumHoldFunction freezes (holds) the spectrum display. The scope trace is
	// suspended and no longer updated until the hold is released.
	SpectrumHoldFunction Function = "SPECTRUM_HOLD"
	// SquelchFunction enables the squelch. When enabled the receiver audio is muted until
	// a signal exceeds the squelch threshold.
	SquelchFunction Function = "SQL"
	// SyncFunction enables frequency synchronization between VFOs. Keeps two VFOs locked
	// to the same frequency, useful for certain operating techniques.
	SyncFunction Function = "SYNC"
	// ToneBurstFunction enables a 1750 Hz tone burst. Used in Europe to open repeaters
	// that require a short 1750 Hz access tone instead of CTCSS.
	ToneBurstFunction Function = "TBURST"
	// ToneFunction enables CTCSS tone transmission. When active, the rig encodes a
	// sub-audible CTCSS tone onto the transmitted signal to access tone-squelched repeaters.
	ToneFunction Function = "TONE"
	// ToneSquelchFunction enables CTCSS tone squelch for receive. The receiver audio is
	// only unmuted when an incoming signal carries the configured CTCSS tone.
	ToneSquelchFunction Function = "TSQL"
	// TransceiveFunction enables transceive mode. When active, the rig pushes unsolicited
	// status updates (frequency, mode changes) to the client. See also SetTransceive.
	TransceiveFunction Function = "TRANSCEIVE"
	// TunerFunction enables the automatic antenna tuner. When toggled on, the rig activates
	// its built-in ATU to match the antenna impedance.
	TunerFunction Function = "TUNER"
	// VoiceSquelchControlFunction enables Voice Squelch Control (VSC). Opens the squelch
	// only when actual voice modulation is detected, not just carrier.
	VoiceSquelchControlFunction Function = "VSC"
	// VOXFunction enables VOX (Voice Operated Transmission). The rig automatically switches
	// to transmit when voice audio is detected on the microphone input.
	VOXFunction Function = "VOX"
	// XITFunction enables XIT (Transmitter Incremental Tuning). Activates the XIT circuit,
	// allowing the transmit frequency to be offset from the receive frequency.
	XITFunction Function = "XIT"
)

// Level represents an analog/graduated rig setting that has a continuous or stepped value
// range, as opposed to functions which are simple on/off toggles. Levels include controls
// like AF gain, RF gain, squelch, RF power, microphone gain, and CW pitch, as well as
// read-only meter values like signal strength (STRENGTH), SWR, ALC, and compression meter
// (COMP_METER). Many writable levels use a normalized 0.0 to 1.0 range, while read-only
// meter levels and some settings (e.g. CWPITCH in Hz, KEYSPD in WPM) use absolute values.
// Not all rigs support all levels; use [RigClient.GetAvailableLevels] to discover which
// levels the connected rig supports.
type Level string

const (
	// AGCLevel sets the Automatic Gain Control speed. The value is an integer token:
	// 0=OFF, 1=SUPERFAST, 2=FAST, 3=SLOW, 4=USER, 5=MEDIUM, 6=AUTO, 7=LONG, 8=ON.
	// Faster AGC responds more quickly to signal changes; slower AGC gives more
	// stable audio on steady signals.
	AGCLevel Level = "AGC"
	// AGCTimeLevel sets the AGC time constant in milliseconds (absolute value). This is
	// the time the AGC takes to release (recover) after a signal disappears.
	AGCTimeLevel Level = "AGC_TIME"
	// ALCLevel is a read-only meter showing the current ALC (Automatic Level Control)
	// reading. ALC limits the drive to the final amplifier stage; consistently high
	// ALC indicates over-driving the transmitter.
	ALCLevel Level = "ALC"
	// AntiVOXLevel sets the anti-VOX sensitivity (normalized 0.0–1.0). Anti-VOX prevents
	// the receiver audio from falsely triggering VOX during transmission.
	AntiVOXLevel Level = "ANTIVOX"
	// AttenuatorLevel sets the RF attenuator value in dB (absolute value). A positive value
	// inserts attenuation ahead of the receiver to reduce strong-signal overload.
	// Use GetAvailableLevels to discover the steps supported by the rig (e.g. 6, 12, 18 dB).
	AttenuatorLevel Level = "ATT"
	// AudioFrequencyLevel sets the AF (audio frequency) gain, i.e. the receiver volume
	// (normalized 0.0–1.0).
	AudioFrequencyLevel Level = "AF"
	// AudioPeakFilterLevel sets the Audio Peak Filter center frequency offset in Hz
	// (absolute value). Tunes the APF notch/peak to match the CW pitch.
	AudioPeakFilterLevel Level = "APF"
	// BalanceLevel sets the audio balance between the main and sub receivers
	// (normalized 0.0–1.0). Only meaningful on dual-receiver rigs.
	BalanceLevel Level = "BAL"
	// BandSelectLevel selects the active amateur band. The value is a band name token
	// such as "BAND160M", "BAND80M", "BAND40M", "BAND20M", "BAND2M", "BAND70CM", etc.
	BandSelectLevel Level = "BAND_SELECT"
	// BreakInDelayLevel sets the CW break-in delay in dot lengths (absolute integer value).
	// Controls how long the rig waits after the last element before switching back to
	// receive in semi-break-in mode.
	BreakInDelayLevel Level = "BKINDL"
	// BreakInDelayMSLevel sets the CW break-in delay in milliseconds (absolute integer).
	// An alternative to BreakInDelayLevel that uses milliseconds instead of dot lengths.
	BreakInDelayMSLevel Level = "BKIN_DLYMS"
	// CompressionMeterLevel is a read-only meter showing the amount of speech compression
	// being applied in dB.
	CompressionMeterLevel Level = "COMP_METER"
	// CompressorLevel sets the speech compressor gain (normalized 0.0–1.0). Higher values
	// increase the average transmitted power for improved readability.
	CompressorLevel Level = "COMP"
	// CurrentDrainLevel is a read-only meter showing the current drain (ID) of the
	// final amplifier in amperes.
	CurrentDrainLevel Level = "ID_METER"
	// CWPitchLevel sets the CW sidetone and receive pitch in Hz (absolute value, e.g. 600).
	// This is the audio frequency at which CW signals are heard.
	CWPitchLevel Level = "CWPITCH"
	// IFLevel sets the IF (Intermediate Frequency) shift in Hz (signed absolute value).
	// Shifts the IF passband up or down relative to the carrier, useful for moving
	// interfering signals out of the passband.
	IFLevel Level = "IF"
	// KeyerSpeedLevel sets the CW keyer speed in words per minute (WPM), absolute value.
	KeyerSpeedLevel Level = "KEYSPD"
	// MeterLevel selects which parameter is shown on the rig's analog meter. The value
	// is a string token: "SWR", "COMP", "ALC", "IC" (current), "VDD" (drain voltage), etc.
	// Rig-dependent; not all meters are available on all rigs.
	MeterLevel Level = "METER"
	// MicGainLevel sets the microphone gain (normalized 0.0–1.0).
	MicGainLevel Level = "MICGAIN"
	// MonitorGainLevel sets the monitor (sidetone) gain (normalized 0.0–1.0). Controls
	// the volume of the transmitted audio fed back to the headphones.
	MonitorGainLevel Level = "MONITOR_GAIN"
	// NoiseBlankerLevel sets the noise blanker threshold or depth (normalized 0.0–1.0).
	// Higher values make the noise blanker more aggressive.
	NoiseBlankerLevel Level = "NB"
	// NotchFilterLevel sets the manual notch filter center frequency in Hz (absolute value).
	// Tunes the notch to the frequency of an interfering carrier.
	NotchFilterLevel Level = "NOTCHF"
	// NotchFilterRawLevel sets the notch filter in raw rig units (absolute integer). Use
	// when direct access to the rig's internal notch value is required.
	NotchFilterRawLevel Level = "NOTCHF_RAW"
	// NRLevel sets the DSP noise reduction depth (normalized 0.0–1.0). Higher values apply
	// stronger noise reduction but may introduce audio artifacts.
	NRLevel Level = "NR"
	// PBTInLevel sets the inner (low-frequency edge) passband tuning offset in Hz (signed).
	// Shifts the lower edge of the IF filter passband, narrowing or widening the filter
	// from the low side.
	PBTInLevel Level = "PBT_IN"
	// PBTOutLevel sets the outer (high-frequency edge) passband tuning offset in Hz (signed).
	// Shifts the upper edge of the IF filter passband, narrowing or widening the filter
	// from the high side.
	PBTOutLevel Level = "PBT_OUT"
	// PreampLevel sets the preamplifier level in dB (absolute value). Enables and selects
	// the strength of the receive preamplifier. Use 0 to disable the preamp.
	// Use GetAvailableLevels to discover the steps supported by the rig (e.g. 10, 20 dB).
	PreampLevel Level = "PREAMP"
	// RawStrengthLevel is a read-only meter returning the raw, uncalibrated ADC signal
	// strength value. Useful for diagnostics or rig-specific signal strength calculations.
	RawStrengthLevel Level = "RAWSTR"
	// RFLevel sets the RF gain (normalized 0.0–1.0). Reduces receiver sensitivity to
	// handle strong signals without reducing audio volume (unlike the attenuator).
	RFLevel Level = "RF"
	// RFPowerLevel sets the transmitter output power (normalized 0.0–1.0 where 1.0 is
	// maximum power). Use Power2mW / MW2Power to convert between normalized and milliwatts.
	RFPowerLevel Level = "RFPOWER"
	// RFPowerMeterLevel is a read-only meter returning the transmitted RF power as a
	// normalized value (0.0–1.0).
	RFPowerMeterLevel Level = "RFPOWER_METER"
	// RFPowerMeterWattsLevel is a read-only meter returning the transmitted RF power in
	// watts (absolute value).
	RFPowerMeterWattsLevel Level = "RFPOWER_METER_WATTS"
	// SlopeHighLevel sets the high-frequency slope (upper cutoff) of the DSP filter in Hz
	// (absolute value). Adjusts where the filter rolls off on the high-frequency side.
	SlopeHighLevel Level = "SLOPE_HIGH"
	// SlopeLowLevel sets the low-frequency slope (lower cutoff) of the DSP filter in Hz
	// (absolute value). Adjusts where the filter rolls off on the low-frequency side.
	SlopeLowLevel Level = "SLOPE_LOW"
	// SpectrumAttenuationLevel sets the RF attenuation applied to the spectrum scope input
	// in dB (absolute value).
	SpectrumAttenuationLevel Level = "SPECTRUM_ATT"
	// SpectrumAverageLevel sets the spectrum scope averaging count (absolute integer).
	// Higher values smooth out the display at the cost of slower response.
	SpectrumAverageLevel Level = "SPECTRUM_AVG"
	// SpectrumEdgeHighLevel sets the upper frequency edge of the spectrum scope display
	// in Hz (absolute value). Defines the right boundary of the panadapter view.
	SpectrumEdgeHighLevel Level = "SPECTRUM_EDGE_HIGH"
	// SpectrumEdgeLowLevel sets the lower frequency edge of the spectrum scope display
	// in Hz (absolute value). Defines the left boundary of the panadapter view.
	SpectrumEdgeLowLevel Level = "SPECTRUM_EDGE_LOW"
	// SpectrumModeLevel sets the spectrum scope display mode (absolute integer token).
	// Selects between center mode (scope centered on VFO) and fixed/scroll modes.
	SpectrumModeLevel Level = "SPECTRUM_MODE"
	// SpectrumReferenceLevel sets the reference level (top of the scope display) in dBm
	// (signed absolute value).
	SpectrumReferenceLevel Level = "SPECTRUM_REF"
	// SpectrumSpanLevel sets the frequency span of the spectrum scope display in Hz
	// (absolute value). Wider spans show more of the band; narrower spans show more detail.
	SpectrumSpanLevel Level = "SPECTRUM_SPAN"
	// SpectrumSpeedLevel sets the spectrum scope sweep speed (absolute integer). Higher
	// values update the display faster.
	SpectrumSpeedLevel Level = "SPECTRUM_SPEED"
	// SquelchLevel sets the squelch threshold (normalized 0.0–1.0). Higher values require
	// a stronger signal to open the squelch.
	SquelchLevel Level = "SQL"
	// StrengthLevel is a read-only meter returning the received signal strength in dBm or
	// as a calibrated S-unit value (rig-dependent). This is the standard signal meter.
	StrengthLevel Level = "STRENGTH"
	// SWRLevel is a read-only meter returning the SWR (Standing Wave Ratio) on the
	// antenna. Values near 1.0 indicate a good match; higher values indicate mismatch.
	SWRLevel Level = "SWR"
	// TemperatureMeterLevel is a read-only meter returning the internal temperature of
	// the rig in degrees Celsius.
	TemperatureMeterLevel Level = "TEMP_METER"
	// USBAudioInputLevel sets the gain of the USB audio input (normalized 0.0–1.0) for
	// rigs with a built-in USB audio interface (e.g. Icom IC-7300).
	USBAudioInputLevel Level = "USB_AF_INPUT"
	// USBAudioLevel sets the USB audio output (AF) level (normalized 0.0–1.0) for rigs
	// with a built-in USB audio interface.
	USBAudioLevel Level = "USB_AF"
	// VoltageDrainLevel is a read-only meter returning the drain voltage (VDD) of the
	// final amplifier stage in volts.
	VoltageDrainLevel Level = "VD_METER"
	// VOXDelayLevel sets the VOX hang time (delay before returning to receive) in
	// milliseconds (absolute value).
	VOXDelayLevel Level = "VOXDELAY"
	// VOXGainLevel sets the VOX sensitivity (normalized 0.0–1.0). Higher values make VOX
	// trigger more easily from quieter audio.
	VOXGainLevel Level = "VOXGAIN"
	// VOXLevel sets the VOX trip level (normalized 0.0–1.0). The minimum audio level
	// required to activate VOX.
	VOXLevel Level = "VOX"
)

// Parameter represents a rig-wide setting that is not specific to any particular VFO.
// Unlike levels and functions which are per-VFO, parameters apply globally to the entire
// rig. Examples include auto power-off timeout (APO), backlight brightness (BACKLIGHT),
// beep on/off (BEEP), and battery level (BAT). Not all rigs support all parameters; use
// [RigClient.GetAvailableParameters] to discover which parameters the connected rig supports.
type Parameter string

const (
	// AFIFOutputACCParm selects the ACC (accessory connector) socket as the AF/IF output
	// destination. Routes the received audio or IF signal to the rear ACC port.
	AFIFOutputACCParm Parameter = "AFIF_ACC"
	// AFIFOutputLANParm selects the LAN (network) interface as the AF/IF output
	// destination. Routes the audio/IF stream over the network.
	AFIFOutputLANParm Parameter = "AFIF_LAN"
	// AFIFOutputParm selects the AF/IF output destination. The accepted values are
	// rig-dependent; use GetParm with "?" to list available options.
	AFIFOutputParm Parameter = "AFIF"
	// AFIFOutputWLANParm selects the WLAN (wireless network) interface as the AF/IF
	// output destination.
	AFIFOutputWLANParm Parameter = "AFIF_WLAN"
	// AnnouncerParm controls the rig's voice announcer (ANN). The value is typically
	// an integer selecting the announcer mode (e.g. 0=off, 1=English, 2=Japanese).
	AnnouncerParm Parameter = "ANN"
	// AutoPowerOffParm sets the auto power-off (APO) timeout in minutes. The rig will
	// power itself off automatically after the specified period of inactivity.
	// Use 0 to disable auto power-off.
	AutoPowerOffParm Parameter = "APO"
	// BacklightParm sets the brightness of the rig's display backlight (normalized 0.0–1.0).
	BacklightParm Parameter = "BACKLIGHT"
	// BandselectParm selects the active band. The accepted value is a band name token
	// such as "BAND160M", "BAND80M", "BAND40M", "BAND20M", "BAND2M", "BAND70CM", etc.
	BandselectParm Parameter = "BANDSELECT"
	// BatteryLevelParm is a read-only parameter returning the current battery charge level
	// (normalized 0.0–1.0) on battery-powered rigs.
	BatteryLevelParm Parameter = "BAT"
	// BeepParm enables or disables the rig's key-press beep (1=on, 0=off).
	BeepParm Parameter = "BEEP"
	// KeyerTypeParm selects the type of CW keyer. Accepted values are rig-dependent;
	// common options include "STRAIGHT" (straight key), "BUG", "PADDLE", "AUTO".
	// Use GetParm with "?" to list the options supported by the connected rig.
	KeyerTypeParm Parameter = "KEYERTYPE"
	// KeylightParm sets the brightness of the rig's key/button backlight
	// (normalized 0.0–1.0).
	KeylightParm Parameter = "KEYLIGHT"
	// ScreensaverParm sets the screensaver timeout in minutes. The rig's display will
	// dim or activate a screensaver after the specified period of inactivity.
	// Use 0 to disable the screensaver.
	ScreensaverParm Parameter = "SCREENSAVER"
	// TimeParm sets or gets the rig's real-time clock as a time string. See also
	// SetClock / GetClock for a typed alternative.
	TimeParm Parameter = "TIME"
)

// PTTStatus represents the Push-To-Talk state of the rig, indicating whether the rig
// is transmitting and through which input path. PTTOff means the rig is receiving.
// PTTOn indicates transmitting without specifying the source. PTTOnMic and PTTOnData
// specify that transmission is routed through the microphone or data port, respectively.
type PTTStatus int

const (
	// PTTOff means the rig is not transmitting (receiving). Pass this to SetPTT to
	// stop transmission and return to receive mode.
	PTTOff PTTStatus = iota
	// PTTOn activates transmission without specifying the audio source. The rig uses
	// whichever input is currently configured (microphone, data port, etc.).
	PTTOn
	// PTTOnMic activates transmission and routes the microphone input as the audio
	// source. Use this to ensure voice audio is transmitted regardless of any data
	// port configuration.
	PTTOnMic
	// PTTOnData activates transmission and routes the data port (rear audio input) as
	// the audio source. Use this for digital mode operation (e.g. FT8, PSK31) to ensure
	// the rig uses the sound card audio rather than the microphone.
	PTTOnData
)

// PowerStatus represents the power state of the rig. The values are bitmask-based:
// PowerOff (0), PowerOn (1), PowerStandby (2), PowerOperate (4), PowerUnknown (8).
// Note that PowerOn and PowerOperate are distinct: some rigs power on into a standby
// state and must be switched to operate mode separately.
type PowerStatus int

const (
	// PowerOn means the rig is powered on (value 1). Some rigs power on directly into
	// operate mode, while others require a separate transition to PowerOperate.
	PowerOn PowerStatus = 1 << iota
	// PowerStandby means the rig is in standby mode (value 2). The rig is powered but
	// not ready to transmit. Common on Icom rigs that boot into standby before switching
	// to operate mode.
	PowerStandby
	// PowerOperate means the rig is in operate mode (value 4). The rig is fully powered
	// and ready to transmit and receive. Use SetPowerStatus with PowerOperate to bring
	// a rig out of standby.
	PowerOperate
	// PowerUnknown means the rig's power state cannot be determined (value 8).
	PowerUnknown
	// PowerOff means the rig is powered off (value 0). Only a small set of commands
	// (GetPowerStatus, SetPowerStatus, DumpCaps, DumpState) can be executed while off.
	PowerOff PowerStatus = 0
)

// ResetMode represents the type of reset to perform on the rig. The values are
// bitmask-based: ResetNone (0), ResetSoft (1) for a software reset, ResetVFO (2)
// to reset VFO settings, ResetMCall (4) to reset the memory call channel, and
// ResetMaster (8) for a full factory reset.
type ResetMode int

const (
	// ResetSoft performs a software reset (value 1). Restarts the rig's internal firmware
	// or microcontroller without a full power cycle. The extent of the reset is rig-dependent.
	ResetSoft ResetMode = 1 << iota
	// ResetVFO resets all VFO settings to defaults (value 2). Clears VFO frequencies,
	// modes, and related parameters without affecting memory channels or global configuration.
	ResetVFO
	// ResetMCall resets the memory call (CALL) channel (value 4). Clears the special
	// call-channel memory that many rigs provide for quickly recalling a specific frequency.
	ResetMCall
	// ResetMaster performs a full master reset (value 8), equivalent to a factory reset.
	// All memories, configuration, and settings are cleared and restored to factory defaults.
	// Use with caution as this action is irreversible.
	ResetMaster
	// ResetNone performs no reset (value 0).
	ResetNone ResetMode = 0
)

// TransceiveMode controls how the rig reports state changes to the client.
// TransceiveOff disables automatic notifications. TransceiveRig enables rig-initiated
// notifications where the rig pushes state changes to the client asynchronously.
// TransceivePoll enables poll-based notification where hamlib periodically queries
// the rig for changes.
type TransceiveMode int

const (
	// TransceiveOff disables transceive mode. The client must explicitly poll the rig
	// for frequency, mode, and other state changes.
	TransceiveOff TransceiveMode = iota
	// TransceiveRig enables rig-initiated transceive. The rig hardware autonomously sends
	// unsolicited status packets over the serial/network connection when its state changes
	// (e.g. operator tunes the VFO). Requires rig support for autonomous notifications.
	TransceiveRig
	// TransceivePoll enables poll-based transceive. Hamlib periodically queries the rig
	// for state changes and forwards them as if they were unsolicited. Use this on rigs
	// that do not support autonomous notifications but where you still want event-driven
	// updates.
	TransceivePoll
)

// Uplink selects which VFO is used for the uplink (transmit) path during satellite
// operation. UplinkNone (0) means no uplink assignment, UplinkSub (1) uses the Sub VFO,
// and UplinkMain (2) uses the Main VFO for the uplink.
type Uplink int

const (
	// UplinkNone disables uplink VFO assignment (value 0). No uplink correction is applied.
	UplinkNone = iota
	// UplinkSub assigns the Sub VFO as the uplink (transmit) path (value 1). Hamlib will
	// apply Doppler compensation on the Sub VFO frequency during satellite passes.
	UplinkSub
	// UplinkMain assigns the Main VFO as the uplink (transmit) path (value 2). Hamlib will
	// apply Doppler compensation on the Main VFO frequency during satellite passes.
	UplinkMain
)

// VFOOp represents a VFO operation that can be performed on the rig. These operations
// include copying or exchanging VFO contents, transferring data between VFO and memory,
// stepping frequency up/down, changing bands, starting the antenna tuner, and toggling
// between VFO A and B.
type VFOOp string

const (
	// CopyVFO copies the current VFO's frequency and mode to the other VFO. After this
	// operation both VFOs have the same settings.
	CopyVFO VFOOp = "CPY"
	// ExchangeVFO swaps the frequency and mode of VFO A and VFO B. Useful for instantly
	// switching to a previously set frequency while preserving the current one.
	ExchangeVFO VFOOp = "XCHG"
	// MemoryFromVFO transfers the current VFO settings (frequency, mode, etc.) into the
	// currently selected memory channel.
	MemoryFromVFO VFOOp = "FROM_VFO"
	// MemoryToVFO copies the currently selected memory channel's settings into the VFO.
	MemoryToVFO VFOOp = "TO_VFO"
	// ClearMemory erases the contents of the currently selected memory channel.
	ClearMemory VFOOp = "MCL"
	// VFOUp increments the VFO frequency by one tuning step (as set by SetTuningStep).
	VFOUp VFOOp = "UP"
	// VFODown decrements the VFO frequency by one tuning step (as set by SetTuningStep).
	VFODown VFOOp = "DOWN"
	// VFOBandUp moves to the next higher amateur band, changing frequency and mode
	// according to the rig's band-change logic.
	VFOBandUp VFOOp = "BAND_UP"
	// VFOBandDown moves to the next lower amateur band, changing frequency and mode
	// according to the rig's band-change logic.
	VFOBandDown VFOOp = "BAND_DOWN"
	// VFOLeft moves the VFO or selection to the left. The exact behaviour is
	// rig-dependent (e.g. selecting the previous memory channel or a UI navigation action).
	VFOLeft VFOOp = "LEFT"
	// VFORight moves the VFO or selection to the right. The exact behaviour is
	// rig-dependent (e.g. selecting the next memory channel or a UI navigation action).
	VFORight VFOOp = "RIGHT"
	// TuneVFO activates the rig's built-in antenna tuner and starts the tuning cycle.
	// The rig briefly transmits a carrier and adjusts the ATU for minimum SWR.
	TuneVFO VFOOp = "TUNE"
	// ToggleVFO toggles the active VFO between VFO A and VFO B, equivalent to pressing
	// the A/B button on the rig's front panel.
	ToggleVFO VFOOp = "TOGGLE"
)

// ScanFunction represents the type of scan operation to perform. StopScan halts any
// active scan. ScanMemory scans through memory channels. ScanSelected scans only
// selected/tagged memory channels. ScanPrio monitors a priority channel. ScanProg
// scans a programmed frequency range. ScanDelta scans a range around the current
// frequency (delta-f scan). ScanVFO scans the VFO frequency range. ScanPipeline
// is a priority channel scan variant.
type ScanFunction string

const (
	// StopScan stops any currently active scan and returns the rig to normal operation.
	StopScan ScanFunction = "STOP"
	// ScanMemory scans sequentially through all memory channels, stopping on each active
	// channel when a signal opens the squelch.
	ScanMemory ScanFunction = "MEM"
	// ScanSelected scans only the memory channels that have been tagged/selected by the
	// operator, skipping untagged channels.
	ScanSelected ScanFunction = "SLCT"
	// ScanPrio monitors a designated priority channel. The rig periodically checks the
	// priority channel for activity while operating normally. The scanChannel argument
	// specifies the priority channel number.
	ScanPrio ScanFunction = "PRIO"
	// ScanProg performs a programmed frequency scan between two stored frequency limits.F
	// The rig steps through a frequency range, pausing when the squelch opens.
	ScanProg ScanFunction = "PROG"
	// ScanDelta performs a delta-f scan centered on the current frequency. The rig scans
	// a range of ±scanChannel Hz around the current VFO frequency.
	ScanDelta ScanFunction = "DELTA"
	// ScanVFO scans the VFO frequency range stored as VFO scan limits on the rig.
	ScanVFO ScanFunction = "VFO"
	// ScanPipeline is a priority scan variant (pipeline mode) where the rig alternates
	// between monitoring multiple channels in a round-robin fashion.
	ScanPipeline ScanFunction = "PLT"
)

// RigInfo combines the most important information about the rig's current state in one struct.
type RigInfo struct {
	VFOs          []RigInfoVFO
	SplitActive   bool
	SATModeActive bool
	Rig           string
	App           string
	Version       string
	Model         string
	CRC           string
}

// RigInfoVFO contains the current state of one VFO as part of RigInfo.
type RigInfoVFO struct {
	VFO       VFO
	Frequency Frequency
	Mode      Mode
	Passband  Bandwidth
	RXActive  bool
	TXActive  bool
}
