package hl

type Request struct {
	Command string
	Args    []string
}

type Response struct {
	CommandEcho string
	Data        map[string]string
	ReturnCode  ReturnCode
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
