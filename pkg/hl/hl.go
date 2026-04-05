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
	result, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("cannot parse %s value: %w", key, err)
	}
	return result, nil
}

func (r Response) GetFloat64(key string) (float64, error) {
	value, ok := r.Data[key]
	if !ok {
		return 0, fmt.Errorf("no %s value", key)
	}
	result, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("cannot parse %s value: %w", key, err)
	}
	return result, nil
}

func (r Response) GetBool(key string) (bool, error) {
	value, ok := r.Data[key]
	if !ok {
		return false, fmt.Errorf("no %s value", key)
	}
	result, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("cannot parse %s value: %w", key, err)
	}
	return result, nil
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
	VFOA     VFO = "VFOA"
	VFOB     VFO = "VFOB"
	VFOC     VFO = "VFOC"
	CurrVFO  VFO = "currVFO"
	MainVFO  VFO = "Main"
	SubVFO   VFO = "Sub"
	TXVFO    VFO = "TX"
	RXVFO    VFO = "RX"
	MainAVFO VFO = "MainA"
	MainBVFO VFO = "MainB"
	SubAVFO  VFO = "SubA"
	SubBVFO  VFO = "SubB"
)

type Frequency = hamradio.Frequency

type Mode string

const (
	ModeUSB     Mode = "USB"
	ModeLSB     Mode = "LSB"
	ModeCW      Mode = "CW"
	ModeCWR     Mode = "CWR"
	ModeRTTY    Mode = "RTTY"
	ModeRTTYR   Mode = "RTTYR"
	ModeAM      Mode = "AM"
	ModeFM      Mode = "FM"
	ModeWFM     Mode = "WFM"
	ModeAMS     Mode = "AMS"
	ModePKTLSB  Mode = "PKTLSB"
	ModePKTUSB  Mode = "PKTUSB"
	ModePKTFM   Mode = "PKTFM"
	ModeECSSUSB Mode = "ECSSUSB"
	ModeECSSLSB Mode = "ECSSLSB"
	ModeFA      Mode = "FA"
	ModeSAM     Mode = "SAM"
	ModeSAL     Mode = "SAL"
	ModeSAH     Mode = "SAH"
	ModeDSB     Mode = "DSB"
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
