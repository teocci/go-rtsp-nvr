package session

import (
	"path/filepath"
)

const (
	DefaultStaticDirPath = "web"
)

// Server struct
type Server struct {
	Web  WebServer
	RTSP RTSPServer
}

type WebServer struct {
	Enable             bool     `json:"enable"`
	Debug              bool     `json:"debug"`
	LogLevel           int      `json:"log_level"`
	Login              string   `json:"login"`
	Password           string   `json:"password"`
	Path               string   `json:"path"`
	Port               string   `json:"port"`
	HTTPSEnable        bool     `json:"https_enable"`
	HTTPSPort          string   `json:"https_port"`
	HTTPSCert          string   `json:"https_cert"`
	HTTPSKey           string   `json:"https_key"`
	HTTPSAutoTLSEnable bool     `json:"https_auto_tls"`
	HTTPSAutoTLSName   string   `json:"https_auto_tls_name"`
	ICEServers         []string `json:"ice_servers"`
	ICEUsername        string   `json:"ice_username"`
	ICECredential      string   `json:"ice_credential"`
	Token              Token    `json:"token,omitempty"`
	WebRTCPortMin      uint16   `json:"webrtc_port_min"`
	WebRTCPortMax      uint16   `json:"webrtc_port_max"`
}

type RTSPServer struct {
	Port string `json:"port"`
}

// Token authorization
type Token struct {
	Enable  bool   `json:"enable,omitempty"`
	Backend string `json:"backend,omitempty"`
}

// WebPath path to the static directory
func (s *Session) WebPath() (path string) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	path = filepath.Clean(s.Server.Web.Path)
	if path == "." {
		return DefaultStaticDirPath
	}

	return
}

// WebDebug read debug options
func (s *Session) WebDebug() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.Server.Web.Debug
}

// WebLogLevel read debug options
func (s *Session) WebLogLevel() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.Server.Web.LogLevel
}

// WebEnable read demo options
func (s *Session) WebEnable() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.Server.Web.Enable
}

// WebLogin read Login options
func (s *Session) WebLogin() string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.Server.Web.Login
}

// WebPassword read Password options
func (s *Session) WebPassword() string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.Server.Web.Password
}

// WebPort read HTTP Port options
func (s *Session) WebPort() string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.Server.Web.Port
}

// RTSPPort read HTTP Port options
func (s *Session) RTSPPort() string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.Server.RTSP.Port
}

// HTTPSEnable if https protocol is enabled
func (s *Session) HTTPSEnable() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.Server.Web.HTTPSEnable
}

// HTTPSPort read HTTPS Port options
func (s *Session) HTTPSPort() string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.Server.Web.HTTPSPort
}

// HTTPSAutoTLSEnable read HTTPS Port options
func (s *Session) HTTPSAutoTLSEnable() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.Server.Web.HTTPSAutoTLSEnable
}

// HTTPSAutoTLSName read HTTPS Port options
func (s *Session) HTTPSAutoTLSName() string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.Server.Web.HTTPSAutoTLSName
}

// ServerHTTPSCert read HTTPS Cert options
func (s *Session) ServerHTTPSCert() string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.Server.Web.HTTPSCert
}

// HTTPSKey read HTTPS Key options
func (s *Session) HTTPSKey() string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.Server.Web.HTTPSKey
}

// ICEServers read ICE servers
func (s *Session) ICEServers() []string {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.Server.Web.ICEServers
}

// ICEUsername read ICE username
func (s *Session) ICEUsername() string {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.Server.Web.ICEUsername
}

// ICECredential read ICE credential
func (s *Session) ICECredential() string {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	return s.Server.Web.ICECredential
}

// WebTokenEnable read HTTPS Key options
func (s *Session) WebTokenEnable() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.Server.Web.Token.Enable
}

// WebTokenBackend read HTTPS Key options
func (s *Session) WebTokenBackend() string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.Server.Web.Token.Backend
}

// WebRTCPortMin read WebRTC Port Min
func (s *Session) WebRTCPortMin() uint16 {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.Server.Web.WebRTCPortMin
}

// WebRTCPortMax read WebRTC Port Max
func (s *Session) WebRTCPortMax() uint16 {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.Server.Web.WebRTCPortMax
}
