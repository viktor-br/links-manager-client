package main

import "path/filepath"

// Config represents configuration parameters.
type Config struct {
	Dir                 string
	AuthTokenFilename   string
	CredentialsFilename string
	APIHost             string
	LogFilename         string
	StorageName         string
}

// CredentialsPath returns path to the credentials file
func (config *Config) CredentialsPath() string {
	return config.Dir + string(filepath.Separator) + config.CredentialsFilename
}

// AuthTokenPath returns path to the authentication toke file
func (config *Config) AuthTokenPath() string {
	return config.Dir + string(filepath.Separator) + config.AuthTokenFilename
}

// LogPath returns path to the authentication toke file
func (config *Config) LogPath() string {
	return config.Dir + string(filepath.Separator) + config.LogFilename
}

// StoragePath returns path to the authentication toke file
func (config *Config) StoragePath() string {
	return config.Dir + string(filepath.Separator) + config.StorageName
}
