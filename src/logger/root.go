/*
 * ************************** BEGIN LICENSE BLOCK ******************************
 *
 * Copyright © 2023 Christian Stolze
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * ************************** END LICENSE BLOCK ********************************
 */

package logger

/*

The logger package handles the logging of infos, debug, and error messages. It
if mainly important for development and bug tracing capabilities of the Renderhive
Service app.

*/

import (

    // standard
    "fmt"
    "os"
    "io/ioutil"
    "time"
    "strings"
	  "path/filepath"

    // external
    "github.com/rs/zerolog"
    "github.com/rs/zerolog/log"

    // internal
    . "renderhive/globals"

)


// structure for the main and package loggers
type LoggerManager struct {

  // Directories
  WorkingDirectory string

  // Writers
  FileWriter *os.File
  ConsoleWriter zerolog.ConsoleWriter

  // Loggers
  Main *zerolog.Logger
  Package map[string]*zerolog.Logger

}

// LOGGER MANAGER
// #############################################################################
// create the instance for the logger manager to be used in all packages
var Manager *LoggerManager

// Initialize everything required for the IPFS management
func (logm *LoggerManager) Init() (error) {
    var err error

    // set global logger level
    // TODO: This can interfere with other packages that use zerolog
    zerolog.SetGlobalLevel(COMPILER_RENDERHIVE_LOGGER_LEVEL)

    // create a file writer with a log file in the log directory
    logm.WorkingDirectory, err = os.Getwd()
    if err != nil {
      log.Error().Err(err).Msg("There was an error creating a temporary file four our log.")
    }
    logm.FileWriter, err = ioutil.TempFile(filepath.Join(logm.WorkingDirectory, "tmp"), "renderhive_service_*.log")
    if err != nil {
      // Can we log an error before we have our logger? :)
      log.Error().Err(err).Msg("There was an error creating a temporary file four our log.")
    }

    // create a console writer
    logm.ConsoleWriter = zerolog.ConsoleWriter{
      Out: os.Stdout,
      TimeFormat: time.RFC822,
      FormatFieldName: func(i interface{}) string {
          return ""
      },
      FormatFieldValue: func(i interface{}) string {
          return ""//strings.ToUpper(fmt.Sprintf("(%s)", i))
      },
      FormatCaller: func(i interface{}) string {
          parts := strings.Split(fmt.Sprintf("%s", i), string(os.PathSeparator))
          filepath := strings.Join(parts[len(parts)-2:], string(os.PathSeparator))
          return fmt.Sprintf("%s", filepath)
      },
    }

    // create the main logger with a multi-output: to a log file andthe console
    MainLogger := zerolog.New(zerolog.MultiLevelWriter(logm.FileWriter, logm.ConsoleWriter)).Level(zerolog.DebugLevel).With().Timestamp().Caller().Str("module", "renderhive").Logger()
    logm.Main = &MainLogger

    // create the package logger map
    logm.Package = make(map[string]*zerolog.Logger)

    // add the package loggers
    logm.AddPackageLogger("logger")
    logm.AddPackageLogger("node")
    logm.AddPackageLogger("hedera")
    logm.AddPackageLogger("ipfs")
    logm.AddPackageLogger("renderer")
    logm.AddPackageLogger("webapp")
    logm.AddPackageLogger("cli")

    // assign this manager to the global manager variable, which will be used
    // by all other packages to call the loggers
    Manager = logm

    return err

}

// Deinitialize the logger manager
func (logm *LoggerManager) DeInit() (error) {
    var err error

    // log debug event
    logm.Package["logger"].Debug().Msg("Deinitializing the logger manager ...")

    return err

}

// add a new logger for a package of the app
func (logm *LoggerManager) AddPackageLogger(name string) *zerolog.Logger {

  // create a new package logger and add it to the global structure map
  PackageLogger := logm.Main.With().Str("package", name).Caller().Logger()
  logm.Package[name] = &PackageLogger

  return logm.Package[name]
}
