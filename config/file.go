package config

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"gopkg.in/yaml.v2"
)

const (
	defaultStatusEndpoint = "/duty/status"
	defaultConfigFile     = "./duty.yaml"

	configFileEnv = "DUTY_CONFIG_FILE"
)

// File represents all the configurable options of Duty
type File struct {
	Routes    []Route           `yaml:"routes"`
	routesMap map[string]*Route `yaml:"-"`
	Status    string            `yaml:"status"`
}

// ParseFromFile reads an Duty config file from the file specified in the
// environment or from the default file location if no environment is specified.
// A File with the populated values is returned and any errors encountered while
// trying to read the file.
func ParseFromFile() (*File, error) {
	configFile := os.Getenv(configFileEnv)

	if configFile == "" {
		configFile = defaultConfigFile
	}

	b, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("Failed to read config file: %v", err.Error())
	}

	var conf File
	err = yaml.Unmarshal(b, &conf)
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal config file: %v", err.Error())
	}

	if conf.Status == "" {
		conf.Status = defaultStatusEndpoint
	}

	conf.routesMap = make(map[string]*Route)
	for i, r := range conf.Routes {
		conf.routesMap[r.Endpoint] = &conf.Routes[i]
	}

	return &conf, nil
}

func (f *File) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == f.Status {
		handleStatus(w, r)
		return
	}

	route, found := f.getRoute(r.URL)
	if !found {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("path not found"))
		return
	}

	route.ServeHTTP(w, r)
}

func (f *File) getRoute(reqURL *url.URL) (*Route, bool) {
	r, found := f.routesMap[reqURL.Path]
	if !found {
		return nil, false
	}

	return r, true
}

func handleStatus(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("duty is functioning"))
}