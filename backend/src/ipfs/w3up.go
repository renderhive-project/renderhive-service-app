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

package ipfs

/*

The w3up service (future service of web3.storage) is used as the filecoin
pinning service by the Renderhive. It will make sure that certain files
remain available to the network in case the file owner goes offline. This,
for example, includes .blend files of render jobs that were submitted by a
node that subsequently goes online.

*/

import (

	// standard

	// external

	// internal
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"renderhive/logger"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	//. "renderhive/globals"
)

// W3UP CLI, AGENT, SPACES, etc.
// #############################################################################
// Agent and CLI data of w3up command line interface
type w3cliAgent struct {

	// w3cli info
	Path    string // Path to the w3 CLI
	Version string // Build version of this CLI instance

	// Process status
	Cmd     *exec.Cmd     // pointer to the exec.Command type
	Param   []string      // Command line options this process was called with
	PID     int           // PID of the process
	Running bool          // Is the process still running
	StdOut  io.ReadCloser // Command-line standard output of the Blender app
	StdErr  io.ReadCloser // Command-line error output of the Blender app

	// Agent data
	DIDkey string
	Email  string

	// Agent spaces
	Spaces      []w3cliSpace
	ActiveSpace int

	// Agent delegations and proofs
	Delegations []w3cliUCAN
	Proofs      []w3cliUCAN
}

// Space data
type w3cliSpace struct {

	// Space information
	Name   string // Name of this space
	DIDkey string // DID key of this space
	Email  string // Email this space is associated with

	// Uploads
	Uploads []w3cliUpload
}

// Uploads in a space
type w3cliUpload struct {

	// Upload information
	Root   string   `json:"root"`
	Shards []string `json:"shards"`
}

// UCAN data
type w3cliUCAN struct {

	// UCAN
	CID          string       `json:"cid"`
	Issuer       string       `json:"issuer"`
	Audience     string       `json:"audience"`
	Capabilities []Capability `json:"capabilities"`
}

type Capability struct {
	With string `json:"with"`
	Can  string `json:"can"`
}

// W3UP FUNCTIONS
// #############################################################################
// Function to process CLI output
func (w3cli *w3cliAgent) ProcessOutput(name string, output io.ReadCloser) error {
	var err error
	var line string

	// Create a new scanner to read the command output
	scanner := bufio.NewScanner(output)
	for scanner.Scan() {

		// read the line
		line = scanner.Text()

		// BASIC FUNCTIONS
		// ***********************************************************************
		// if the agent data was requested
		if strings.EqualFold(w3cli.Cmd.Args[1], "whoami") {

			// Compile regular expression to match the values
			didkeytest := regexp.MustCompile("did:key")

			// if the expression matches
			if didkeytest.MatchString(line) {

				// extract the values
				w3cli.DIDkey = line

			}

			// if a list of all files in the current space was requested
		} else if strings.EqualFold(w3cli.Cmd.Args[1], "ls") {

			// Compile regular expression to match the JSON listing of uploads
			jsonkeytest := regexp.MustCompile(`^\{"root":"ba[a-z0-9]+","shards":\["ba[a-z0-9]+"(,"ba[a-z0-9]+")*\]\}$`)

			// if the expression matches
			if jsonkeytest.MatchString(strings.ReplaceAll(line, " ", "")) && (len(w3cli.Spaces) > 0 && w3cli.ActiveSpace < len(w3cli.Spaces)) {

				// Current space
				currentSpace := &w3cli.Spaces[w3cli.ActiveSpace]

				// Unmarshal the JSON data into the UCAN struct
				var Upload w3cliUpload
				if err := json.Unmarshal([]byte(line), &Upload); err != nil {
					fmt.Println("Error:", err)
					return err
				}
				currentSpace.Uploads = append(currentSpace.Uploads, Upload)

			}

		}

		// SPACE MANAGEMENT
		// ***********************************************************************
		// if the space functions were invoked
		if strings.EqualFold(w3cli.Cmd.Args[1], "space") {

			// if a new space was created
			if strings.EqualFold(w3cli.Cmd.Args[2], "create") {

				// Compile regular expression to match the values
				didkeytest := regexp.MustCompile("did:key")

				// if the expression matches
				if didkeytest.MatchString(line) {

					// Add the space data to the internal data
					w3cli.Spaces = append(w3cli.Spaces, w3cliSpace{
						Name:   w3cli.Cmd.Args[3],
						DIDkey: line,
					})

					// Make this space the current space
					w3cli.ActiveSpace = len(w3cli.Spaces) - 1

				}

				// if a space was added via a delegation proof
			} else if strings.EqualFold(w3cli.Cmd.Args[2], "add") {

				// Compile regular expression to match the values
				didkeytest := regexp.MustCompile("did:key")

				// if the expression matches
				if didkeytest.MatchString(line) {

					// reset the status variable
					found := false

					// loop through the internal data structure
					for i := range w3cli.Spaces {
						if w3cli.Spaces[i].DIDkey == line {

							// set the status variable
							found = true

							// stop the loop
							break

						}
					}

					// Add the space data to the internal data, if it doesn't exist
					if found == false {
						w3cli.Spaces = append(w3cli.Spaces, w3cliSpace{
							DIDkey: line,
						})
					}

					// Make this space the current space
					w3cli.ActiveSpace = len(w3cli.Spaces) - 1

				}

				// if a space was registered
			} else if strings.EqualFold(w3cli.Cmd.Args[2], "register") {

				// add the email address to tge internal space data
				w3cli.Spaces[w3cli.ActiveSpace].Email = w3cli.Email

				// if the active space was changed
			} else if strings.EqualFold(w3cli.Cmd.Args[2], "use") {

				// Compile regular expression to match the values
				didkeytest := regexp.MustCompile("did:key")

				// if the expression matches
				if didkeytest.MatchString(line) {

					for i := range w3cli.Spaces {
						if w3cli.Spaces[i].DIDkey == line {

							// Make this space the current space
							w3cli.ActiveSpace = i
							break

						}
					}

				}

				// if the spaces are listed
			} else if strings.EqualFold(w3cli.Cmd.Args[2], "ls") {

				// Compile regular expression to match the values
				didkeytest := regexp.MustCompile("did:key")

				// if the expression matches
				line_parts := strings.Split(line, " ")
				if didkeytest.MatchString(line_parts[len(line_parts)-2]) {

					// reset the status variable
					found := false

					// loop through the internal data structure
					for i := range w3cli.Spaces {
						if w3cli.Spaces[i].DIDkey == line_parts[len(line_parts)-2] {

							// mark this space as the current space
							if line_parts[len(line_parts)-3] == "*" {
								w3cli.ActiveSpace = i
							}

							// set the status variable
							found = true

							// stop the loop
							break

						}
					}

					// Add the space data to the internal data, if it doesn't exist
					if found == false {
						w3cli.Spaces = append(w3cli.Spaces, w3cliSpace{
							Name:   line_parts[len(line_parts)-1],
							DIDkey: line_parts[len(line_parts)-2],
						})
					}
				}

			}

		}

		// CAPABILITY MANAGEMENT
		// ***********************************************************************
		// if the 'delegation' functions were invoked
		if strings.EqualFold(w3cli.Cmd.Args[1], "delegation") {

			// if a delegation was created
			if strings.EqualFold(w3cli.Cmd.Args[2], "create") {

				// if the expression matches
				if line != "" {

					// TODO: PROCESS THE BINARY BLOB IN CAR FORMAT HERE

				}

				// if the delegations of this agent shall be listed
			} else if strings.EqualFold(w3cli.Cmd.Args[2], "ls") {

				// if a non-empty line was retrieved
				if line != "" {

					// Unmarshal the JSON data into the UCAN struct
					var UCAN w3cliUCAN
					if err := json.Unmarshal([]byte(line), &UCAN); err != nil {
						fmt.Println("Error:", err)
						return err
					}
					w3cli.Delegations = append(w3cli.Delegations, UCAN)
				}

			}

			// if the 'proof' functions were invoked
		} else if strings.EqualFold(w3cli.Cmd.Args[1], "proof") {

			// if a proof was added
			if strings.EqualFold(w3cli.Cmd.Args[2], "add") {

				// if the expression matches
				if line != "" {

					// TODO: PROCESS THE OUTPUT

				}

				// if the delegations of this agent shall be listed
			} else if strings.EqualFold(w3cli.Cmd.Args[2], "ls") {

				// if a non-empty line was retrieved
				if line != "" {

					// Unmarshal the JSON data into the UCAN struct
					var UCAN w3cliUCAN
					if err := json.Unmarshal([]byte(line), &UCAN); err != nil {
						fmt.Println("Error:", err)
						return err
					}
					w3cli.Proofs = append(w3cli.Proofs, UCAN)

				}

			}
		}

	}

	return err

}

// BASIC FUNCTIONS
// ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

// Authorize this node as an w3up agent for the user's email address
func (w3cli *w3cliAgent) Authorize(email string) error {
	var err error

	// Get the DID string of the agent
	_, err = w3cli.Whoami()
	if err != nil {
		return err
	}

	// if there is NO did key for a w3up agent
	if w3cli.DIDkey == "" {

		// execute the corresponding command line interface call
		w3cli.Cmd = exec.Command(w3cli.Path, "authorize", email)
		w3cli.StdOut, _ = w3cli.Cmd.StdoutPipe()
		w3cli.StdErr, _ = w3cli.Cmd.StderrPipe()
		err = w3cli.Cmd.Start()
		if err != nil {
			return err
		}

		// check of error output
		go w3cli.ProcessOutput("StdErr", w3cli.StdErr)

		// check for output
		err = w3cli.ProcessOutput("StdOut", w3cli.StdOut)
		if err != nil {
			return err
		}
	}

	// if a DID key for the w3up exists
	if w3cli.DIDkey != "" {

		logger.Manager.Package["ipfs"].Info().Msg(fmt.Sprintf(" [#] Agent DID (w3up): %v", w3cli.DIDkey))

	}

	return err

}

// Query the agent DID
func (w3cli *w3cliAgent) Whoami() (string, error) {
	var err error

	// execute the corresponding command line interface call
	w3cli.Cmd = exec.Command(w3cli.Path, "whoami")
	w3cli.StdOut, _ = w3cli.Cmd.StdoutPipe()
	w3cli.StdErr, _ = w3cli.Cmd.StderrPipe()
	err = w3cli.Cmd.Start()
	if err != nil {
		return "", err
	}

	// check of error output
	go w3cli.ProcessOutput("StdErr", w3cli.StdErr)

	// check for output
	err = w3cli.ProcessOutput("StdOut", w3cli.StdOut)
	if err != nil {
		return "", err
	}

	return w3cli.DIDkey, err

}

// Upload the given file(s) to the w3up service
func (w3cli *w3cliAgent) Upload(paths []string) error {
	var err error

	// Only proceed, if the agent DID is known
	if w3cli.DIDkey == "" {
		return fmt.Errorf("This w3up agent seems to be not initialized.")
	}

	// execute the corresponding command line interface call
	w3cli.Cmd = exec.Command(w3cli.Path, append([]string{"up"}, paths...)...)
	w3cli.StdOut, _ = w3cli.Cmd.StdoutPipe()
	w3cli.StdErr, _ = w3cli.Cmd.StderrPipe()
	err = w3cli.Cmd.Start()
	if err != nil {
		return err
	}

	// check of error output
	go w3cli.ProcessOutput("StdErr", w3cli.StdErr)

	// check for output
	err = w3cli.ProcessOutput("StdOut", w3cli.StdOut)
	if err != nil {
		return err
	}

	return err

}

// Remove an upload from the uploads listing.
// NOTE: This  does not remove the data from the IPFS network, nor does it remove it from space storage (by default).
func (w3cli *w3cliAgent) Remove(cid string) error {
	var err error

	// Only proceed, if the agent DID is known
	if w3cli.DIDkey == "" {
		return fmt.Errorf("This w3up agent seems to be not initialized.")
	}

	// execute the corresponding command line interface call
	w3cli.Cmd = exec.Command(w3cli.Path, "rm", cid)
	w3cli.StdOut, _ = w3cli.Cmd.StdoutPipe()
	w3cli.StdErr, _ = w3cli.Cmd.StderrPipe()
	err = w3cli.Cmd.Start()
	if err != nil {
		return err
	}

	// check of error output
	go w3cli.ProcessOutput("StdErr", w3cli.StdErr)

	// check for output
	err = w3cli.ProcessOutput("StdOut", w3cli.StdOut)
	if err != nil {
		return err
	}

	return err

}

// List all the uploads registered in the current space.
func (w3cli *w3cliAgent) UploadList() error {
	var err error

	// Only proceed, if the agent DID is known AND there are spaces
	if w3cli.DIDkey == "" || len(w3cli.Spaces) > 0 {
		return fmt.Errorf("This w3up agent seems to be not initialized.")
	}

	// Empty the list of uploads
	w3cli.Spaces[w3cli.ActiveSpace].Uploads = nil

	// execute the corresponding command line interface call
	w3cli.Cmd = exec.Command(w3cli.Path, "ls", "--json")
	w3cli.StdOut, _ = w3cli.Cmd.StdoutPipe()
	w3cli.StdErr, _ = w3cli.Cmd.StderrPipe()
	err = w3cli.Cmd.Start()
	if err != nil {
		return err
	}

	// check of error output
	go w3cli.ProcessOutput("StdErr", w3cli.StdErr)

	// wait for the output
	err = w3cli.ProcessOutput("StdOut", w3cli.StdOut)
	if err != nil {
		return err
	}

	return err

}

// SPACE MANAGEMENT
// ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

// Create a space for uploads in the w3up service
func (w3cli *w3cliAgent) SpaceCreate(name string) error {
	var err error

	// Only proceed, if the agent DID is known
	if w3cli.DIDkey == "" {
		return fmt.Errorf("This w3up agent seems to be not initialized.")
	}

	// execute the corresponding command line interface call
	w3cli.Cmd = exec.Command(w3cli.Path, "space", "create", name)
	w3cli.StdOut, _ = w3cli.Cmd.StdoutPipe()
	w3cli.StdErr, _ = w3cli.Cmd.StderrPipe()
	err = w3cli.Cmd.Start()
	if err != nil {
		return err
	}

	// check of error output
	go w3cli.ProcessOutput("StdErr", w3cli.StdErr)

	// check for output
	err = w3cli.ProcessOutput("StdOut", w3cli.StdOut)
	if err != nil {
		return err
	}

	return err

}

// Add a space for uploads in the w3up service to this agent
// The proof is a CAR encoded delegation to _this_ agent.
func (w3cli *w3cliAgent) SpaceAdd(proof string) error {
	var err error

	// Only proceed, if the agent DID is known
	if w3cli.DIDkey == "" {
		return fmt.Errorf("This w3up agent seems to be not initialized.")
	}

	// execute the corresponding command line interface call
	w3cli.Cmd = exec.Command(w3cli.Path, "space", "add", proof)
	w3cli.StdOut, _ = w3cli.Cmd.StdoutPipe()
	w3cli.StdErr, _ = w3cli.Cmd.StderrPipe()
	err = w3cli.Cmd.Start()
	if err != nil {
		return err
	}

	// check of error output
	go w3cli.ProcessOutput("StdErr", w3cli.StdErr)

	// check for output
	err = w3cli.ProcessOutput("StdOut", w3cli.StdOut)
	if err != nil {
		return err
	}

	// // Add the space data to the internal data
	// space = w3cliSpace{
	// 	Name: name,
	// 	DIDkey: name
	// }
	// w3cli.Spaces = append(w3cli.Spaces, space)

	return err

}

// Register a space with an account in the w3up service
func (w3cli *w3cliAgent) SpaceRegister() error {
	var err error

	// Only proceed, if the agent DID is known AND there are spaces
	if w3cli.DIDkey == "" || len(w3cli.Spaces) > 0 {
		return fmt.Errorf("This w3up agent seems to be not initialized.")
	}

	// execute the corresponding command line interface call
	w3cli.Cmd = exec.Command(w3cli.Path, "space", "register", "--email", w3cli.Email, "--provider", "did:web:web3.storage")
	w3cli.StdOut, _ = w3cli.Cmd.StdoutPipe()
	w3cli.StdErr, _ = w3cli.Cmd.StderrPipe()
	err = w3cli.Cmd.Start()
	if err != nil {
		return err
	}

	// check of error output
	go w3cli.ProcessOutput("StdErr", w3cli.StdErr)

	// check for output
	err = w3cli.ProcessOutput("StdOut", w3cli.StdOut)
	if err != nil {
		return err
	}

	return err

}

// Set the active space by DID
func (w3cli *w3cliAgent) SpaceUse(did string) error {
	var err error

	// Only proceed, if the agent DID is known AND there are spaces
	if w3cli.DIDkey == "" || len(w3cli.Spaces) > 0 {
		return fmt.Errorf("This w3up agent seems to be not initialized.")
	}

	// execute the corresponding command line interface call
	w3cli.Cmd = exec.Command(w3cli.Path, "space", "use", did)
	w3cli.StdOut, _ = w3cli.Cmd.StdoutPipe()
	w3cli.StdErr, _ = w3cli.Cmd.StderrPipe()
	err = w3cli.Cmd.Start()
	if err != nil {
		return err
	}

	// check of error output
	go w3cli.ProcessOutput("StdErr", w3cli.StdErr)

	// check for output
	err = w3cli.ProcessOutput("StdOut", w3cli.StdOut)
	if err != nil {
		return err
	}

	return err

}

// Request a list of spaces known by this agent
func (w3cli *w3cliAgent) SpaceList() error {
	var err error

	// Only proceed, if the agent DID is known
	if w3cli.DIDkey == "" {
		return fmt.Errorf("This w3up agent seems to be not initialized.")
	}

	// execute the corresponding command line interface call
	w3cli.Cmd = exec.Command(w3cli.Path, "space", "ls")
	w3cli.StdOut, _ = w3cli.Cmd.StdoutPipe()
	w3cli.StdErr, _ = w3cli.Cmd.StderrPipe()
	err = w3cli.Cmd.Start()
	if err != nil {
		return err
	}

	// check of error output
	go w3cli.ProcessOutput("StdErr", w3cli.StdErr)

	// check for output
	err = w3cli.ProcessOutput("StdOut", w3cli.StdOut)
	if err != nil {
		return err
	}

	// fmt.Print(w3cli.Spaces)

	return err

}

// CAPABILITIES MANAGEMENT
// ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

// Create a delegation to the passed audience DID for the given abilities with the current space as the resource
func (w3cli *w3cliAgent) DelegationCreate(did string, capabilities []string, expiration_time int64, audiencetype string, proof_path string) error {
	var err error
	var args []string

	// Only proceed, if the agent DID is known
	if w3cli.DIDkey == "" {
		return fmt.Errorf("This w3up agent seems to be not initialized.")
	}

	// construct the complete argument chain for the CLI call
	// first the capabilities
	for i := range capabilities {
		args = append(args, []string{"--can", "'" + capabilities[i] + "'"}...)
	}

	// then the options
	args = append(args, []string{
		"--output", proof_path,
		"--expiration", strconv.FormatInt(expiration_time, 10),
		"--type", audiencetype,
	}...)

	// execute the corresponding command line interface call
	w3cli.Cmd = exec.Command(w3cli.Path, append([]string{"delegation", "create", did}, args...)...)
	w3cli.StdOut, _ = w3cli.Cmd.StdoutPipe()
	w3cli.StdErr, _ = w3cli.Cmd.StderrPipe()
	err = w3cli.Cmd.Start()
	if err != nil {
		return err
	}

	// check of error output
	go w3cli.ProcessOutput("StdErr", w3cli.StdErr)

	// check for output
	err = w3cli.ProcessOutput("StdOut", w3cli.StdOut)
	if err != nil {
		return err
	}

	return err

}

// List delegations created by this agent for others
func (w3cli *w3cliAgent) DelegationList() error {
	var err error

	// Only proceed, if the agent DID is known
	if w3cli.DIDkey == "" {
		return fmt.Errorf("This w3up agent seems to be not initialized.")
	}

	// execute the corresponding command line interface call
	w3cli.Cmd = exec.Command(w3cli.Path, "delegation", "ls", "--json")
	w3cli.StdOut, _ = w3cli.Cmd.StdoutPipe()
	w3cli.StdErr, _ = w3cli.Cmd.StderrPipe()
	err = w3cli.Cmd.Start()
	if err != nil {
		return err
	}

	// check of error output
	go w3cli.ProcessOutput("StdErr", w3cli.StdErr)

	// check for output
	err = w3cli.ProcessOutput("StdOut", w3cli.StdOut)
	if err != nil {
		return err
	}

	return err

}

// Add a proof delegated to this agent. The proof is a CAR encoded delegation to this agent
// NOTE: Use w3 'space add' unless you know the delegation you received targets a resource other than a w3 space.
func (w3cli *w3cliAgent) ProofAdd(proof_path string) error {
	var err error

	// Only proceed, if the agent DID is known
	if w3cli.DIDkey == "" {
		return fmt.Errorf("This w3up agent seems to be not initialized.")
	}

	// execute the corresponding command line interface call
	w3cli.Cmd = exec.Command(w3cli.Path, "proof", "add", proof_path)
	w3cli.StdOut, _ = w3cli.Cmd.StdoutPipe()
	w3cli.StdErr, _ = w3cli.Cmd.StderrPipe()
	err = w3cli.Cmd.Start()
	if err != nil {
		return err
	}

	// check of error output
	go w3cli.ProcessOutput("StdErr", w3cli.StdErr)

	// check for output
	err = w3cli.ProcessOutput("StdOut", w3cli.StdOut)
	if err != nil {
		return err
	}

	return err

}

// List delegations created by this agent for others
func (w3cli *w3cliAgent) ProofList() error {
	var err error

	// Only proceed, if the agent DID is known
	if w3cli.DIDkey == "" {
		return fmt.Errorf("This w3up agent seems to be not initialized.")
	}

	// execute the corresponding command line interface call
	w3cli.Cmd = exec.Command(w3cli.Path, "proof", "ls", "--json")
	w3cli.StdOut, _ = w3cli.Cmd.StdoutPipe()
	w3cli.StdErr, _ = w3cli.Cmd.StderrPipe()
	err = w3cli.Cmd.Start()
	if err != nil {
		return err
	}

	// check of error output
	go w3cli.ProcessOutput("StdErr", w3cli.StdErr)

	// check for output
	err = w3cli.ProcessOutput("StdOut", w3cli.StdOut)
	if err != nil {
		return err
	}

	return err

}

// COMMAND LINE INTERFACE - W3 UP SERVICE
// #############################################################################
// Create the CLI command to interact with the w3up service
func (ipfsm *PackageManager) CreateCommandW3() *cobra.Command {

	// create a 'w3' command for the node
	command := &cobra.Command{
		Use:   "w3",
		Short: "Interact with the Filecoin-based w3up service",
		Long:  "This command the base command to interact with the Filecoin layer of the renderhive.",
		Run: func(cmd *cobra.Command, args []string) {

			// if the w3 agent is NOT initialized
			if ipfsm.W3Agent.DIDkey == "" {

				fmt.Println("")
				fmt.Println(fmt.Errorf("Could not find a w3 agent for this node."))
				fmt.Println("")
			}

			return

		},
	}

	// add the subcommands
	command.AddCommand(ipfsm.CreateCommandW3_Info())

	return command

}

// List all information about the w3 up agent of this node
func (ipfsm *PackageManager) CreateCommandW3_Info() *cobra.Command {
	var err error

	// create a 'info' command for the w3 agent
	command := &cobra.Command{
		Use:   "info",
		Short: "Print information about the w3up service agent",
		Long:  "This command prints the available information of the w3up service agent.",
		Run: func(cmd *cobra.Command, args []string) {

			// if the w3 agent is initialized
			if ipfsm.W3Agent.DIDkey != "" {

				// Print info
				fmt.Println("")
				fmt.Println(fmt.Sprintf("The w3up service agent information:"))

				// Authorize this agent
				_, err = ipfsm.W3Agent.Whoami()
				if err != nil {
					fmt.Println(fmt.Errorf(err.Error()))
					return
				}

				// Print info
				fmt.Println(fmt.Sprintf(" [#] DID: %v", ipfsm.W3Agent.DIDkey))

				// Get list of spaces this agent has access to
				err = ipfsm.W3Agent.SpaceList()
				if err != nil {
					fmt.Println(fmt.Errorf(err.Error()))
					return
				}

				// Print info
				fmt.Println(fmt.Sprintf(" [#] Number of available spaces: %v", len(ipfsm.W3Agent.Spaces)))

				// if there is at least one space
				if len(ipfsm.W3Agent.Spaces) > 0 {

					// Print info
					fmt.Println(fmt.Sprintf(" [#] Current space: '%v' (%v)", ipfsm.W3Agent.Spaces[ipfsm.W3Agent.ActiveSpace].Name, ipfsm.W3Agent.ActiveSpace))

					// Get list of uploads in the active space
					err = ipfsm.W3Agent.UploadList()
					if err != nil {
						fmt.Println(fmt.Errorf(err.Error()))
						return
					}

					// Print info
					fmt.Println(fmt.Sprintf(" [#] Number of available files in current space: %v", len(ipfsm.W3Agent.Spaces[ipfsm.W3Agent.ActiveSpace].Uploads)))

				}

			} else {

				fmt.Println("")
				fmt.Println(fmt.Errorf("Could not find a w3 agent for this node."))
				fmt.Println("")

			}

			return

		},
	}

	return command

}
