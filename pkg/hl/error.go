package hl

import "fmt"

// RigError is a custom error type that contains a ReturnCode and a human-readable error message.
type RigError struct {
	Code    ReturnCode
	Message string
}

// Error implements the error interface for RigError.
func (e *RigError) Error() string {
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// Error variables for all ReturnCode constants
var (
	ErrInvalidParameter          = &RigError{RigInvalidParameter, "invalid parameter"}
	ErrInvalidConfiguration      = &RigError{RigInvalidConfiguration, "invalid configuration"}
	ErrMemoryShortage            = &RigError{RigMemoryShortage, "memory shortage"}
	ErrFeatureNotImplemented     = &RigError{RigFeatureNotImplemented, "feature not implemented"}
	ErrCommunicationTimeout      = &RigError{RigCommunicationTimeout, "communication timed out"}
	ErrIOError                   = &RigError{RigIOError, "IO error"}
	ErrInternalError             = &RigError{RigInternalError, "internal Hamlib error"}
	ErrProtocolError             = &RigError{RigProtocolError, "protocol error"}
	ErrCommandRejected           = &RigError{RigCommandRejected, "command rejected by the rig"}
	ErrArgumentTruncated         = &RigError{RigArgumentTruncated, "command performed, but arg truncated, result not guaranteed"}
	ErrFeatureNotAvailable       = &RigError{RigFeatureNotAvailable, "feature not available"}
	ErrVFONotAccessible          = &RigError{RigVFONotAccessible, "target VFO unaccessible"}
	ErrCommunicationBusError     = &RigError{RigCommunicationBusError, "communication bus error"}
	ErrCommunicationBusCollision = &RigError{RigCommunicationBusCollision, "communication bus collision"}
	ErrNullHandle                = &RigError{RigNullHandle, "NULL RIG handle or invalid pointer parameter"}
	ErrInvalidVFO                = &RigError{RigInvalidVFO, "invalid VFO"}
	ErrArgumentOutOfDomain       = &RigError{RigArgumentOutOfDomain, "argument out of domain of func"}
	ErrFunctionDeprecated        = &RigError{RigFunctionDeprecated, "function deprecated"}
	ErrSecurityError             = &RigError{RigSecurityError, "security error password not provided or crypto failure"}
	ErrNotPoweredOn              = &RigError{RigNotPoweredOn, "rig is not powered on"}
	ErrLimitExceeded             = &RigError{RigLimitExceeded, "limit exceeded"}
	ErrAccessDenied              = &RigError{RigAccessDenied, "access denied"}
)

var returnCodeMap = map[ReturnCode]*RigError{
	RigInvalidParameter:          ErrInvalidParameter,
	RigInvalidConfiguration:      ErrInvalidConfiguration,
	RigMemoryShortage:            ErrMemoryShortage,
	RigFeatureNotImplemented:     ErrFeatureNotImplemented,
	RigCommunicationTimeout:      ErrCommunicationTimeout,
	RigIOError:                   ErrIOError,
	RigInternalError:             ErrInternalError,
	RigProtocolError:             ErrProtocolError,
	RigCommandRejected:           ErrCommandRejected,
	RigArgumentTruncated:         ErrArgumentTruncated,
	RigFeatureNotAvailable:       ErrFeatureNotAvailable,
	RigVFONotAccessible:          ErrVFONotAccessible,
	RigCommunicationBusError:     ErrCommunicationBusError,
	RigCommunicationBusCollision: ErrCommunicationBusCollision,
	RigNullHandle:                ErrNullHandle,
	RigInvalidVFO:                ErrInvalidVFO,
	RigArgumentOutOfDomain:       ErrArgumentOutOfDomain,
	RigFunctionDeprecated:        ErrFunctionDeprecated,
	RigSecurityError:             ErrSecurityError,
	RigNotPoweredOn:              ErrNotPoweredOn,
	RigLimitExceeded:             ErrLimitExceeded,
	RigAccessDenied:              ErrAccessDenied,
}

// ReturnCodeAsError converts a ReturnCode value into the corresponding error variable.
func ReturnCodeAsError(code ReturnCode) error {
	if code == RigOk {
		return nil
	}
	if err, ok := returnCodeMap[code]; ok {
		return err
	}
	return &RigError{code, "unknown return code"}
}
