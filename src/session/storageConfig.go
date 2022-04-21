package session

import (
	"encoding/json"
	"github.com/hashicorp/go-version"
	"github.com/liip/sheriff"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
)

var CoreSession = New()

const defaultConfigFilePath = "config.json"

//debug global
var debug bool

//New loads the config file
func New() *Session {
	//argConfigPatch := flag.String("config", "config.json", "config patch (/etc/server/config.json or config.json)")
	//argDebug := flag.Bool("debug", true, "set debug Mode")
	//debug = *argDebug
	//flag.Parse()
	cwdPath, _ := os.Getwd()
	configPath := path.Join(filepath.Dir(filepath.Dir(cwdPath)), defaultConfigFilePath)
	var session *Session
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Println("ReadFile", err)
		os.Exit(1)
	}

	err = json.Unmarshal(data, &session)
	if err != nil {
		log.Println("Unmarshal", err)
		os.Exit(1)
	}

	debug = session.Server.Web.Debug

	session.loadStreams()

	return session
}

func (s *Session) loadStreams() {
	for sessionID, stream := range s.Streams {
		stream.LoadChannels()
		s.Streams[sessionID] = stream
	}
}

//SaveConfig Client
func (s *Session) SaveConfig() (err error) {
	v2, err := version.NewVersion("2.0.0")
	if err != nil {
		return
	}

	o := &sheriff.Options{
		Groups:     []string{"config"},
		ApiVersion: v2,
	}

	data, err := sheriff.Marshal(o, s)
	if err != nil {
		return
	}

	res, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return
	}

	err = ioutil.WriteFile("config.json", res, 0644)
	if err != nil {
		return
	}

	return nil
}
