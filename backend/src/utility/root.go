/*
 * ************************** BEGIN LICENSE BLOCK ******************************
 *
 * Copyright © 2024 Christian Stolze
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

package utility

/*

This package contains utility/helper functions for the service app.

*/

import (

	// standard
	// "fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	// "time"
	// "sync"
	// external
	// hederasdk "github.com/hashgraph/hedera-sdk-go/v2"

	// internal
	. "renderhive/globals"
	// "renderhive/logger"
	// "renderhive/node"
	// "renderhive/hedera"
	// "renderhive/ipfs"
	// "renderhive/jsonrpc"
	// "renderhive/cli"
)

// FUNCTIONS
// #############################################################################
// Initialize the Renderhive Service App session
func InStringSlice(slice []string, test string) bool {

	// loop through all elements and check if one element is the 'test' string
	for _, str := range slice {
		if str == test {
			return true
		}
	}

	return false

}

// Get the app data path as a string
func GetAppDataPath() string {

	// OS-specific path to app data
	app_data_path, err := os.UserConfigDir()
	if err != nil {
		return ""
	}

	return filepath.Join(app_data_path, RENDERHIVE_APP_DIRECTORY)

}

// Query the public IPv4 address of this computer from an external service
func GetPublicIPv4() (string, error) {

	// Make a GET request to an external service that returns the public IP address
	resp, err := http.Get("https://api.ipify.org")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Read the response body and print the IP address
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil

}

// Query the public IPv6 address of this computer from an external service
func GetPublicIPv6() (string, error) {

	// Make a GET request to an external service that returns the public IP address
	resp, err := http.Get("https://api64.ipify.org")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Read the response body and print the IP address
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil

}

// check if a given path is a file (i.e., if it exists and is no directory)
func IsFile(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return !info.IsDir(), nil
}
