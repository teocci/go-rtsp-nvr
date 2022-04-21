// Package config
// Created by RTT.
// Author: teocci@yandex.com on 2021-Aug-23
package config

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/teocci/go-rtsp-nvr/src/logger"
)

const (
	llName  = "log-level"
	llShort = "l"
	llDesc  = "Log level to output [fatal|error|info|debug|trace]"

	lfName  = "log-file"
	lfShort = "L"
	lfDesc  = "File that will receive the logger output"

	sysName  = "syslog"
	sysShort = "S"
	sysDesc  = "Logger will output into the syslog"

	vName  = "verbose"
	vShort = "v"
	vDesc  = "Run in verbose mode"

	cfName  = "config-file"
	cfShort = "c"
	cfDesc  = "Configuration file to load"

	tdName  = "temp-dir"
	tdShort = "t"
	tdDesc  = "Temporal directory where the app will work"

	verName  = "version"
	verShort = "V"
	verDesc  = "Print version info and exit"
)

var (
	Syslog  = false // Run logger will output into the syslog
	Verbose = true  // Run in verbose mode

	File    = ""      // Configuration file to load
	TempDir = "./tmp" // Temporal directory
	Version = false   // Print version info and exit

	LogLevel  = "info"         // Log level to output [fatal|error|info|debug|trace]
	LogFile   = ""             // File where the logger will output
	LogConfig logger.LogConfig // configuration for the logger
	Log       *logger.Logger   // Central logger for the app
)

// AddFlags adds the available cli flags
func AddFlags(cmd *cobra.Command) {
	// core
	cmd.Flags().StringVarP(&LogLevel, llName, llShort, LogLevel, llDesc)
	cmd.Flags().StringVarP(&LogFile, lfName, lfShort, LogFile, lfDesc)
	cmd.Flags().BoolVarP(&Verbose, vName, vShort, Verbose, vDesc)
	cmd.Flags().BoolVarP(&Verbose, sysName, sysShort, Verbose, sysDesc)

	cmd.PersistentFlags().StringVarP(&File, cfName, cfShort, File, cfDesc)
	cmd.PersistentFlags().StringVarP(&TempDir, tdName, tdShort, TempDir, tdDesc)

	cmd.Flags().BoolVarP(&Version, verName, verShort, Version, verDesc)
}

// LoadConfigFile reads the specified config file
func LoadConfigFile() error {
	if File == "" {
		return nil
	}

	// Set defaults to whatever might be there already
	viper.SetDefault(llName, LogLevel)
	viper.SetDefault(vName, Verbose)
	viper.SetDefault(tdName, TempDir)

	filename := filepath.Base(File)
	viper.SetConfigName(filename[:len(filename)-len(filepath.Ext(filename))])
	viper.AddConfigPath(filepath.Dir(File))

	err := viper.ReadInConfig()
	if err != nil {
		return fmt.Errorf("failed to read config file - %v", err)
	}

	// Set values. Config file will override commandline
	LogLevel = viper.GetString(llName)
	Verbose = viper.GetBool(vName)
	TempDir = viper.GetString(tdName)

	LoadLogConfig()

	return nil
}

func LoadLogConfig() {
	LogConfig = logger.LogConfig{
		Level:   LogLevel,
		Verbose: Verbose,
		Syslog:  Syslog,
		LogFile: LogFile,
	}
}
