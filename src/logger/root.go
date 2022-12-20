/*
 * ************************** BEGIN LICENSE BLOCK ******************************
 *
 * Copyright © 2022 Christian Stolze
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
    . "renderhive/constants"

)

// structure for the main and package loggers
type RenderhiveLoggers struct {
    Main *zerolog.Logger
    Package map[string]*zerolog.Logger
}

// initialize the global logger
var RenderhiveLogger = RenderhiveLoggers{}

// initialize the logger package
func Init() {

  // set global logger level
  // TODO: This can interfere with other packages that use zerolog
  zerolog.SetGlobalLevel(COMPILER_RENDERHIVE_LOGGER_LEVEL)

  // create a file writer with a log file in the log directory
  workingDirectory, err := os.Getwd()
  if err != nil {
    log.Error().Err(err).Msg("There was an error creating a temporary file four our log.")
  }
  fileWriter, err := ioutil.TempFile(filepath.Join(workingDirectory, "tmp"), "renderhive_service_*.log")
  if err != nil {
    // Can we log an error before we have our logger? :)
    log.Error().Err(err).Msg("There was an error creating a temporary file four our log.")
  }

  // create a console writer
  consoleWriter := zerolog.ConsoleWriter{
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
  mainLogger := zerolog.New(zerolog.MultiLevelWriter(fileWriter, consoleWriter)).Level(zerolog.DebugLevel).With().Timestamp().Caller().Str("module", "renderhive").Logger()
  RenderhiveLogger.Main = &mainLogger

  fmt.Printf("The log file is allocated at %s\n", fileWriter.Name())

  // create the package logger map
  RenderhiveLogger.Package = make(map[string]*zerolog.Logger)

}

// add a new logger for a package of the app
func AddPackageLogger(name string) *zerolog.Logger {

  // create a new package logger
  NewPackageLogger := RenderhiveLogger.Main.With().Str("package", name).Caller().Logger()

  // add a package logger to the global structure map
  RenderhiveLogger.Package[name] = &NewPackageLogger

  return RenderhiveLogger.Package[name]
}
