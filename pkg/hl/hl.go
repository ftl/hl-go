package hl

import (
	"fmt"
	"slices"
	"strconv"

	"github.com/ftl/hamradio"
)

type Request struct {
	Command string
	Args    []string
}

type Response struct {
	CommandEcho string
	Data        map[string]string
	ReturnCode  ReturnCode
}

func (r Response) GetString(key string) (string, error) {
	result, ok := r.Data[key]
	if !ok {
		return "", fmt.Errorf("no %s value", key)
	}
	return result, nil
}

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

type ReturnCode int

const (
	RigOk ReturnCode = iota * -1
	RigInvalidParameter
	RigInvalidConfiguration
	RigMemoryShortage
	RigFeatureNotImplemented
	RigCommunicationTimeout
	RigIOError
	RigInternalError
	RigProtocolError
	RigCommandRejected
	RigArgumentTruncated
	RigFeatureNotAvailable
	RigVFONotAccessible
	RigCommunicationBusError
	RigCommunicationBusCollision
	RigNullHandle
	RigInvalidVFO
	RigArgumentOutOfDomain
	RigFunctionDeprecated
	RigSecurityError
	RigNotPoweredOn
	RigLimitExceeded
	RigAccessDenied
)

type VFO string

const (
	CurrVFO  VFO = "currVFO"
	MainAVFO VFO = "MainA"
	MainBVFO VFO = "MainB"
	MainVFO  VFO = "Main"
	RXVFO    VFO = "RX"
	SubAVFO  VFO = "SubA"
	SubBVFO  VFO = "SubB"
	SubVFO   VFO = "Sub"
	TXVFO    VFO = "TX"
	VFOA     VFO = "VFOA"
	VFOB     VFO = "VFOB"
	VFOC     VFO = "VFOC"
)

type Frequency = hamradio.Frequency

type Mode string

const (
	ModeAM      Mode = "AM"
	ModeAMS     Mode = "AMS"
	ModeCW      Mode = "CW"
	ModeCWR     Mode = "CWR"
	ModeDSB     Mode = "DSB"
	ModeECSSLSB Mode = "ECSSLSB"
	ModeECSSUSB Mode = "ECSSUSB"
	ModeFA      Mode = "FA"
	ModeFM      Mode = "FM"
	ModeLSB     Mode = "LSB"
	ModePKTFM   Mode = "PKTFM"
	ModePKTLSB  Mode = "PKTLSB"
	ModePKTUSB  Mode = "PKTUSB"
	ModeRTTY    Mode = "RTTY"
	ModeRTTYR   Mode = "RTTYR"
	ModeSAH     Mode = "SAH"
	ModeSAL     Mode = "SAL"
	ModeSAM     Mode = "SAM"
	ModeUSB     Mode = "USB"
	ModeWFM     Mode = "WFM"
)

type ModeBandwidths struct {
	Mode   Mode
	Normal Frequency
	Narrow Frequency
	Wide   Frequency
}

func Modes(m map[Mode]ModeBandwidths) []Mode {
	result := make([]Mode, 0, len(m))
	for mode := range m {
		result = append(result, mode)
	}
	slices.Sort(result)
	return result
}

type Function string

const (
	AdvancedIPFunction           Function = "AIP"
	AudioFilterFunction          Function = "AFLT"
	AudioPeakFilterFunction      Function = "APF"
	AutoBandModeFunction         Function = "ABM"
	AutoFrequencyControlFunction Function = "AFC"
	AutoNoiseLimiterFunction     Function = "ANL"
	AutoNotchFilterFunction      Function = "ANF"
	AutoRepeaterOffsetFunction   Function = "ARO"
	BeatCanceler2Function        Function = "BC2"
	BeatCancellerFunction        Function = "BC"
	CombinedSquelchFunction      Function = "CSQL"
	CompressorFunction           Function = "COMP"
	DigitalSquelchFunction       Function = "DSQL"
	DiversityFunction            Function = "DIVERSITY"
	DualWatchFunction            Function = "DUAL_WATCH"
	FastAGCFunction              Function = "FAGC"
	FullBreakInFunction          Function = "FBKIN"
	LockFunction                 Function = "LOCK"
	ManualBeatCancellerFunction  Function = "MBC"
	ManualNotchFunction          Function = "MN"
	MonitorFunction              Function = "MON"
	MuteFunction                 Function = "MUTE"
	NoiseBlanker2Function        Function = "NB2"
	NoiseBlankerFunction         Function = "NB"
	NoiseReductionFunction       Function = "NR"
	OverflowwStatusFunction      Function = "OVF_STATUS"
	ResumeFunction               Function = "RESUME"
	ReverseFunction              Function = "REV"
	RFFunction                   Function = "RF"
	RITFunction                  Function = "RIT"
	SATModeFunction              Function = "SATMODE"
	SceneFunction                Function = "SCEN"
	ScopeFunction                Function = "SCOPE"
	SemiBreakInFunction          Function = "SBKIN"
	SendMorseFunction            Function = "SEND_MORSE"
	SendVoiceMemoryFunction      Function = "SEND_VOICE_MEM"
	SliceFunction                Function = "SLICE"
	SpectrumFunction             Function = "SPECTRUM"
	SpectrumHoldFunction         Function = "SPECTRUM_HOLD"
	SquelchFunction              Function = "SQL"
	SyncFunction                 Function = "SYNC"
	ToneBurstFunction            Function = "TBURST"
	ToneFunction                 Function = "TONE"
	ToneSquelchFunction          Function = "TSQL"
	TransceiveFunction           Function = "TRANSCEIVE"
	TunerFunction                Function = "TUNER"
	VoiceSquelchControlFunction  Function = "VSC"
	VOXFunction                  Function = "VOX"
	XITFunction                  Function = "XIT"
)

type Level string

const (
	AGCLevel                 Level = "AGC"
	AGCTimeLevel             Level = "AGC_TIME"
	ALCLevel                 Level = "ALC"
	AntiVOXLevel             Level = "ANTIVOX"
	AttenuatorLevel          Level = "ATT"
	AudioFrequencyLevel      Level = "AF"
	AudioPeakFilterLevel     Level = "APF"
	BalanceLevel             Level = "BAL"
	BandSelectLevel          Level = "BAND_SELECT"
	BreakInDelayLevel        Level = "BKINDL"
	BreakInDelayMSLevel      Level = "BKIN_DLYMS"
	CompressionMeterLevel    Level = "COMP_METER"
	CompressorLevel          Level = "COMP"
	CurrentDrainLevel        Level = "ID_METER"
	CWPitchLevel             Level = "CWPITCH"
	IFLevel                  Level = "IF"
	KeyerSpeedLevel          Level = "KEYSPD"
	MeterLevel               Level = "METER"
	MicGainLevel             Level = "MICGAIN"
	MonitorGainLevel         Level = "MONITOR_GAIN"
	NoiseBlankerLevel        Level = "NB"
	NotchFilterLevel         Level = "NOTCHF"
	NotchFilterRawLevel      Level = "NOTCHF_RAW"
	NRLevel                  Level = "NR"
	PBTInLevel               Level = "PBT_IN"
	PBTOutLevel              Level = "PBT_OUT"
	PreampLevel              Level = "PREAMP"
	RawStrengthLevel         Level = "RAWSTR"
	RFLevel                  Level = "RF"
	RFPowerLevel             Level = "RFPOWER"
	RFPowerMeterLevel        Level = "RFPOWER_METER"
	RFPowerMeterWattsLevel   Level = "RFPOWER_METER_WATTS"
	SlopeHighLevel           Level = "SLOPE_HIGH"
	SlopeLowLevel            Level = "SLOPE_LOW"
	SpectrumAttenuationLevel Level = "SPECTRUM_ATT"
	SpectrumAverageLevel     Level = "SPECTRUM_AVG"
	SpectrumEdgeHighLevel    Level = "SPECTRUM_EDGE_HIGH"
	SpectrumEdgeLowLevel     Level = "SPECTRUM_EDGE_LOW"
	SpectrumModeLevel        Level = "SPECTRUM_MODE"
	SpectrumReferenceLevel   Level = "SPECTRUM_REF"
	SpectrumSpanLevel        Level = "SPECTRUM_SPAN"
	SpectrumSpeedLevel       Level = "SPECTRUM_SPEED"
	SquelchLevel             Level = "SQL"
	StrengthLevel            Level = "STRENGTH"
	SWRLevel                 Level = "SWR"
	TemperatureMeterLevel    Level = "TEMP_METER"
	USBAudioInputLevel       Level = "USB_AF_INPUT"
	USBAudioLevel            Level = "USB_AF"
	VoltageDrainLevel        Level = "VD_METER"
	VOXDelayLevel            Level = "VOXDELAY"
	VOXGainLevel             Level = "VOXGAIN"
	VOXLevel                 Level = "VOX"
)

type Parameter string

const (
	AFIFOutputACCParm  Parameter = "AFIF_ACC"
	AFIFOutputLANParm  Parameter = "AFIF_LAN"
	AFIFOutputParm     Parameter = "AFIF"
	AFIFOutputWLANParm Parameter = "AFIF_WLAN"
	AnnouncerParm      Parameter = "ANN"
	AutoPowerOffParm   Parameter = "APO"
	BacklightParm      Parameter = "BACKLIGHT"
	BandselectParm     Parameter = "BANDSELECT"
	BatteryLevelParm   Parameter = "BAT"
	BeepParm           Parameter = "BEEP"
	KeyerTypeParm      Parameter = "KEYERTYPE"
	KeylightParm       Parameter = "KEYLIGHT"
	ScreensaverParm    Parameter = "SCREENSAVER"
	TimeParm           Parameter = "TIME"
)
