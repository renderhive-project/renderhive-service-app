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

package jsonrpc

/*

The jsonrpc package provides the communication layer between backend and frontend for the user UI, which will
be served locally as a JSON-RPC. It is basically a JSON RPC client-server model.

*/

import (

	// standard

	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"

	// "os"
	// "time"

	// external
	"net"
	"net/http"
	"net/http/httptest"

	"github.com/golang-jwt/jwt/v5"

	"github.com/gorilla/mux"
	"github.com/gorilla/rpc/v2"
	"github.com/gorilla/rpc/v2/json2"

	"github.com/spf13/cobra"

	// internal
	"renderhive/logger"
	// "renderhive/globals"
	// "renderhive/hedera"
)

// structure for the JSON-RPC manager
type PackageManager struct {

	// JSON RPC
	// General
	Mutex         sync.Mutex
	JsonRpcServer *rpc.Server
	HttpServer    http.Server
	Listener      net.Listener
	Port          string
	CertFile      string
	KeyFile       string

	// Services
	PingService     *PingService
	ContractService *ContractService
	OperatorService *OperatorService

	// Session data
	SessionActive bool
	SessionToken  struct {
		SignedString string
		ExpiresAt    time.Time
		PublicKey    ed25519.PublicKey
		PrivateKey   ed25519.PrivateKey
		Update       bool
	}
	SessionCookie struct {
		Name string
	}

	// Command line interface
	Command      *cobra.Command
	CommandFlags struct {
		FlagPlaceholder bool
	}
}

// JSON-RPC MANAGER
// #############################################################################
// create the render manager variable
var Manager = PackageManager{}

// Initialize everything required for the JSON-RPC management
func (jsonrpcm *PackageManager) Init() error {
	var err error

	// log information
	logger.Manager.Package["jsonrpc"].Info().Msg("Initializing the JSON-RPC manager ...")

	// Create all services
	jsonrpcm.PingService = new(PingService)
	jsonrpcm.ContractService = new(ContractService)
	jsonrpcm.OperatorService = new(OperatorService)

	return err

}

// Deinitialize the JSON-RPC manager
func (jsonrpcm *PackageManager) DeInit() error {
	var err error

	// log event
	logger.Manager.Package["jsonrpc"].Debug().Msg("Deinitializing the JSON-RPC manager ...")

	return err

}

// Start the JSON-RPC server
func (jsonrpcm *PackageManager) StartServer(port string, certFile string, keyFile string) error {
	var err error

	// Create the RPC server
	jsonrpcm.JsonRpcServer = rpc.NewServer()

	// set JSON codec
	jsonrpcm.JsonRpcServer.RegisterCodec(json2.NewCodec(), "application/json")

	// register all services
	err = jsonrpcm.JsonRpcServer.RegisterService(jsonrpcm.PingService, "PingService")
	if err != nil {
		return err
	}
	err = jsonrpcm.JsonRpcServer.RegisterService(jsonrpcm.ContractService, "ContractService")
	if err != nil {
		return err
	}
	err = jsonrpcm.JsonRpcServer.RegisterService(jsonrpcm.OperatorService, "OperatorService")
	if err != nil {
		return err
	}

	// Create a new router
	router := mux.NewRouter()

	// Apply middleware to the router
	router.Use(jsonrpcm.corsMiddleware)
	router.Use(jsonrpcm.authenticationMiddleware)

	// Handle OPTIONS requests on the JSON-RPC route
	router.HandleFunc("/jsonrpc", func(w http.ResponseWriter, r *http.Request) {
		// log event
		logger.Manager.Package["jsonrpc"].Debug().Msg("Handling the OPTIONS Request")

		// Respond to the OPTIONS request with CORS headers and 200 OK
		w.WriteHeader(http.StatusOK)
		return

	}).Methods("OPTIONS")

	// Define GET method separately for the JSON-RPC route
	router.HandleFunc("/jsonrpc", func(w http.ResponseWriter, r *http.Request) {
		// log event
		logger.Manager.Package["jsonrpc"].Debug().Msg("Handling the GET Request")
		w.Write([]byte("JSON-RPC server active. Please use POST requests for RPC calls."))
	}).Methods("GET")

	// Handle POST requests on the JSON-RPC route
	router.HandleFunc("/jsonrpc", func(w http.ResponseWriter, r *http.Request) {
		// log event
		logger.Manager.Package["jsonrpc"].Debug().Msg("Handling the POST Request")
		// Create a new response writer that buffers the response
		// NOTE: We need this, so that we can write the session cookie after the SignIn request
		bw := NewBufferedResponseWriter(w)

		// Call JSON-RPC method
		jsonrpcm.JsonRpcServer.ServeHTTP(bw, r)

		// if the JWT should be updated
		if jsonrpcm.SessionToken.SignedString != "" && Manager.SessionToken.Update {

			// log event
			logger.Manager.Package["jsonrpc"].Debug().Msg("Setting HttpOnly cookie ...")
			logger.Manager.Package["jsonrpc"].Debug().Msg(fmt.Sprintf(" [#] Name: ", jsonrpcm.SessionCookie.Name))
			logger.Manager.Package["jsonrpc"].Debug().Msg(fmt.Sprintf(" [#] String: ", jsonrpcm.SessionToken.SignedString))

			// set the cookie, which will expire at the same time as the token
			http.SetCookie(w, &http.Cookie{
				Name: jsonrpcm.SessionCookie.Name,
				//Domain:   "localhost",
				Path:     "/",
				Value:    jsonrpcm.SessionToken.SignedString,
				Expires:  jsonrpcm.SessionToken.ExpiresAt,
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteLaxMode,
			})

			// reset update status
			jsonrpcm.SessionToken.Update = false

			// log event
			logger.Manager.Package["jsonrpc"].Info().Msg("Cookie set and user sucessfully logged in.")

		}

		// Write buffered response to the original response writer
		bw.WriteBufferedResponse()

	}).Methods("POST")

	// Setting up HTTPS Server configuration
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}
	jsonrpcm.Port = port
	jsonrpcm.CertFile = certFile
	jsonrpcm.KeyFile = keyFile
	jsonrpcm.HttpServer = http.Server{
		Addr:      ":" + jsonrpcm.Port,
		TLSConfig: tlsConfig,
		Handler:   router,
	}

	// log event
	logger.Manager.Package["jsonrpc"].Debug().Msg(fmt.Sprintf("JSON-RPC server starting on port %v ...", jsonrpcm.Port))

	// Start the server
	err = jsonrpcm.HttpServer.ListenAndServeTLS(jsonrpcm.CertFile, jsonrpcm.KeyFile)
	if err != nil {
		return err
	}

	return err

}

// Stop the JSON RPC server
func (jsonrpcm *PackageManager) StopServer() {

	// log event
	logger.Manager.Package["jsonrpc"].Debug().Msg("Attempting to stop the server.")

	if jsonrpcm.Listener != nil {
		// Create a context with a timeout to allow ongoing requests to complete
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Attempt to gracefully shutdown the server
		if err := jsonrpcm.HttpServer.Shutdown(ctx); err != nil {
			// If the shutdown fails, log the error and force close the listener
			logger.Manager.Package["jsonrpc"].Error().Msgf("Server shutdown failed: %+v", err)
			jsonrpcm.Listener.Close()
		}
	}

	// log event
	logger.Manager.Package["jsonrpc"].Debug().Msg("Server was stopped.")

}

// Redirects a JSON-RPC request in string format to the JSON-RPC internally
// NOTE: This is used to call JSON-RPC methods from within the backend, after receiving a request via Hedera topics (or possibly other channels)
func (jsonrpcm *PackageManager) RedirectToServer(payload string) (string, error) {

	// log event
	logger.Manager.Package["jsonrpc"].Debug().Msg("Calling JSON-RPC method internally ...")
	logger.Manager.Package["jsonrpc"].Debug().Msg(payload)

	// check if the JSON-RPC server is running
	if jsonrpcm.HttpServer.Addr == "" || jsonrpcm.JsonRpcServer == nil {
		return "", fmt.Errorf("JSON-RPC server is not ready")
	}

	// Create a new HTTP request with the JSON-RPC payload as the body
	r, err := http.NewRequest("POST", "/jsonrpc", bytes.NewBuffer([]byte(payload)))
	if err != nil {
		return "", err
	}
	r.Header.Set("Content-Type", "application/json")

	// Create a new ResponseRecorder to capture the response
	w := httptest.NewRecorder()

	// Call the ServeHTTP method of the RPC server, which executes the JSON-RPC method
	jsonrpcm.JsonRpcServer.ServeHTTP(w, r)

	// Extract the response from the ResponseRecorder
	response := w.Body.String()

	return response, nil

}

// JSON-RPC BUFFEREDRESPONSEWRITER
// #############################################################################
type BufferedResponseWriter struct {
	original http.ResponseWriter
	code     int
	body     *bytes.Buffer
	header   http.Header
}

func NewBufferedResponseWriter(original http.ResponseWriter) *BufferedResponseWriter {
	return &BufferedResponseWriter{
		original: original,
		code:     http.StatusOK,
		body:     &bytes.Buffer{},
		header:   make(http.Header),
	}
}

func (b *BufferedResponseWriter) Header() http.Header {
	return b.header
}

func (b *BufferedResponseWriter) WriteHeader(code int) {
	b.code = code
}

func (b *BufferedResponseWriter) Write(p []byte) (int, error) {
	return b.body.Write(p)
}

func (b *BufferedResponseWriter) WriteBufferedResponse() {
	// Write headers
	for key, values := range b.header {
		for _, value := range values {
			b.original.Header().Add(key, value)
		}
	}

	// Write status code
	b.original.WriteHeader(b.code)

	// Write body
	b.original.Write(b.body.Bytes())
}

// JSON-RPC MIDDLEWARE
// #############################################################################

// CORS middleware handler for the router
func (jsonrpcm *PackageManager) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "https://localhost:5173")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, withCredentials")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Handling CORS preflight request
		if r.Method == "OPTIONS" {
			// log event
			logger.Manager.Package["jsonrpc"].Debug().Msg("Handling the OPTIONS Request")

			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Authentication middleware handler for the router
func (jsonrpcm *PackageManager) authenticationMiddleware(next http.Handler) http.Handler {

	// define a set of allowed methods that do not require authentication
	whitelistMethods := map[string]bool{
		"OperatorService.GetSignInPayload": true,
		"OperatorService.GetInfo":          true,
		"OperatorService.SignUp":           true,
		"OperatorService.SignIn":           true,
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// get the method name from the request
		method, err := jsonrpcm.getRpcMethod(w, r)
		if err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
		if _, isAllowed := whitelistMethods[method]; isAllowed {
			next.ServeHTTP(w, r)
			return
		}

		// VERIFY THE JWT
		// Extract JWT session token from HttpOnly cookie
		cookie, err := r.Cookie(jsonrpcm.SessionCookie.Name)
		if err != nil {

			// No cookie, return Unauthorized response
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusUnauthorized)
			return
		}

		// Parse and Verify the JWT seesion token
		tokenString := cookie.Value
		_, err = jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate the algorithm and return the key for verification
			if token.Method != jwt.SigningMethodEdDSA {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}

			return jsonrpcm.SessionToken.PublicKey, nil
		}, jwt.WithValidMethods([]string{"EdDSA"}))
		if err != nil {
			// Invalid token, return Unauthorized response
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusUnauthorized)
			return
		}

		// Call the next handler in the chain
		next.ServeHTTP(w, r)
	})
}

// get the JSON-RPC method name from the HTTP request
func (jsonrpcm *PackageManager) getRpcMethod(w http.ResponseWriter, r *http.Request) (string, error) {

	// Get the name of the requested JSON-RPC method
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return "", err
	}
	r.Body.Close()                                    //  Must close
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // Replace the body

	var requestBody map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &requestBody); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return "", err
	}

	method, ok := requestBody["method"].(string)
	if !ok {
		http.Error(w, "Invalid method", http.StatusBadRequest)
		return "", err
	}

	return method, nil

}

// JSON-RPC MANAGER COMMAND LINE INTERFACE
// #############################################################################
// Create the command for the command line interface
func (jsonrpcm *PackageManager) CreateCommand() *cobra.Command {

	// create the package command
	jsonrpcm.Command = &cobra.Command{
		Use:   "jsonrpc",
		Short: "Commands for the web frontend of the Renderhive Service App",
		Long:  "This command and its sub-commands enable the management of the web frontend for the Renderhive Service App.",
		Run: func(cmd *cobra.Command, args []string) {

			return

		},
	}

	return jsonrpcm.Command

}
