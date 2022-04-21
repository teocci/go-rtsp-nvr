// Package base
// Created by RTT.
// Author: teocci@yandex.com on 2022-Apr-05
package base

// StatusCode is the status code of a RTSP response.
type StatusCode int

// RTSP standard response status codes
const (
	StatusContinue                           StatusCode = 100
	StatusOK                                 StatusCode = 200
	StatusMovedPermanently                   StatusCode = 301
	StatusFound                              StatusCode = 302
	StatusSeeOther                           StatusCode = 303
	StatusNotModified                        StatusCode = 304
	StatusUseProxy                           StatusCode = 305
	StatusBadRequest                         StatusCode = 400
	StatusUnauthorized                       StatusCode = 401
	StatusPaymentRequired                    StatusCode = 402
	StatusForbidden                          StatusCode = 403
	StatusNotFound                           StatusCode = 404
	StatusMethodNotAllowed                   StatusCode = 405
	StatusNotAcceptable                      StatusCode = 406
	StatusProxyAuthRequired                  StatusCode = 407
	StatusRequestTimeout                     StatusCode = 408
	StatusGone                               StatusCode = 410
	StatusPreconditionFailed                 StatusCode = 412
	StatusRequestEntityTooLarge              StatusCode = 413
	StatusRequestURITooLong                  StatusCode = 414
	StatusUnsupportedMediaType               StatusCode = 415
	StatusParameterNotUnderstood             StatusCode = 451
	StatusNotEnoughBandwidth                 StatusCode = 453
	StatusSessionNotFound                    StatusCode = 454
	StatusMethodNotValidInThisState          StatusCode = 455
	StatusHeaderFieldNotValidForResource     StatusCode = 456
	StatusInvalidRange                       StatusCode = 457
	StatusParameterIsReadOnly                StatusCode = 458
	StatusAggregateOperationNotAllowed       StatusCode = 459
	StatusOnlyAggregateOperationAllowed      StatusCode = 460
	StatusUnsupportedTransport               StatusCode = 461
	StatusDestinationUnreachable             StatusCode = 462
	StatusDestinationProhibited              StatusCode = 463
	StatusDataTransportNotReadyYet           StatusCode = 464
	StatusNotificationReasonUnknown          StatusCode = 465
	StatusKeyManagementError                 StatusCode = 466
	StatusConnectionAuthorizationRequired    StatusCode = 470
	StatusConnectionCredentialsNotAccepted   StatusCode = 471
	StatusFailureToEstablishSecureConnection StatusCode = 472
	StatusInternalServerError                StatusCode = 500
	StatusNotImplemented                     StatusCode = 501
	StatusBadGateway                         StatusCode = 502
	StatusServiceUnavailable                 StatusCode = 503
	StatusGatewayTimeout                     StatusCode = 504
	StatusRTSPVersionNotSupported            StatusCode = 505
	StatusOptionNotSupported                 StatusCode = 551
	StatusProxyUnavailable                   StatusCode = 553
)

// StatusMessages contains the status messages associated with each status code.
var StatusMessages = statusMessages

var statusMessages = map[StatusCode]string{
	StatusContinue: "Continue",

	StatusOK: "OK",

	StatusMovedPermanently: "Moved Permanently",
	StatusFound:            "Found",
	StatusSeeOther:         "See Other",
	StatusNotModified:      "Not Modified",
	StatusUseProxy:         "Use Proxy",

	StatusBadRequest:                         "Bad Request",
	StatusUnauthorized:                       "Unauthorized",
	StatusPaymentRequired:                    "Payment Required",
	StatusForbidden:                          "Forbidden",
	StatusNotFound:                           "Not Found",
	StatusMethodNotAllowed:                   "Method Not Allowed",
	StatusNotAcceptable:                      "Not Acceptable",
	StatusProxyAuthRequired:                  "Proxy Auth Required",
	StatusRequestTimeout:                     "Request Timeout",
	StatusGone:                               "Gone",
	StatusPreconditionFailed:                 "Precondition Failed",
	StatusRequestEntityTooLarge:              "Request Entity Too Large",
	StatusRequestURITooLong:                  "Request URI Too Long",
	StatusUnsupportedMediaType:               "Unsupported Media Type",
	StatusParameterNotUnderstood:             "Parameter Not Understood",
	StatusNotEnoughBandwidth:                 "Not Enough Bandwidth",
	StatusSessionNotFound:                    "Session Not Found",
	StatusMethodNotValidInThisState:          "Method Not Valid In This State",
	StatusHeaderFieldNotValidForResource:     "Header Field Not Valid for Resource",
	StatusInvalidRange:                       "Invalid Range",
	StatusParameterIsReadOnly:                "Parameter Is Read-Only",
	StatusAggregateOperationNotAllowed:       "Aggregate Operation Not Allowed",
	StatusOnlyAggregateOperationAllowed:      "Only Aggregate Operation Allowed",
	StatusUnsupportedTransport:               "Unsupported Transport",
	StatusDestinationUnreachable:             "Destination Unreachable",
	StatusDestinationProhibited:              "Destination Prohibited",
	StatusDataTransportNotReadyYet:           "Data Transport Not Ready Yet",
	StatusNotificationReasonUnknown:          "Notification Reason Unknown",
	StatusKeyManagementError:                 "Key Management Error",
	StatusConnectionAuthorizationRequired:    "Connection Authorization Required",
	StatusConnectionCredentialsNotAccepted:   "Connection Credentials Not Accepted",
	StatusFailureToEstablishSecureConnection: "Failure to Establish Secure Connection",

	StatusInternalServerError:     "Internal Server Error",
	StatusNotImplemented:          "Not Implemented",
	StatusBadGateway:              "Bad Gateway",
	StatusServiceUnavailable:      "Service Unavailable",
	StatusGatewayTimeout:          "Gateway Timeout",
	StatusRTSPVersionNotSupported: "RTSP Version Not Supported",
	StatusOptionNotSupported:      "Option Not Supported",
	StatusProxyUnavailable:        "Proxy Unavailable",
}

// Method is the method of a RTSP request.
type Method string

// standard methods
const (
	Announce     Method = "ANNOUNCE"
	Describe     Method = "DESCRIBE"
	GetParameter Method = "GET_PARAMETER"
	Options      Method = "OPTIONS"
	Pause        Method = "PAUSE"
	Play         Method = "PLAY"
	Record       Method = "RECORD"
	Setup        Method = "SETUP"
	SetParameter Method = "SET_PARAMETER"
	Teardown     Method = "TEARDOWN"
)

const (
	rtspProtocol10           = "RTSP/1.0"
	requestMaxMethodLength   = 64
	requestMaxURLLength      = 2048
	requestMaxProtocolLength = 64
)

const (
	TAG             = "[RTSP]"
	Version         = "RTSP/1.0"
	UserAgent       = "Lavf58.29.100"
	Session         = "000a959d6816"
	TransportFormat = "RTP/AVP/TCP;unicast;interleaved=%d-%d"
	SDPContentType  = "application/sdp"
	SeverName       = "gortsplib"
)

const (
	OPTIONS  = "OPTIONS"
	DESCRIBE = "DESCRIBE"
	SETUP    = "SETUP"
	PLAY     = "PLAY"
	TEARDOWN = "TEARDOWN"
)

const (
	HeaderCSeq        = "CSeq"
	HeaderPublic      = "Public"
	HeaderServer      = "Server"
	HeaderSession     = "Session"
	HeaderContentBase = "Content-Base"
	HeaderContentType = "Content-Type"
)

const (
	rtspMaxContentLength = 128 * 1024
)

func (m Method) String() string {
	return string(m)
}

func StatusMessage(code StatusCode) string {
	return statusMessages[code]
}
