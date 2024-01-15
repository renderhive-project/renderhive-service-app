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

package globals

import (

	// standard
	"errors"
	"fmt"
	"time"

	// external
	"github.com/rs/zerolog"
	// internal
	// ...
)

// COMPILER FLAGS
// #############################################################################
const COMPILER_RENDERHIVE_LOGGER_LEVEL = zerolog.DebugLevel

// HEDERA CONSTANTS
// #############################################################################
// Mirror node URL
const HEDERA_TESTNET_MIRROR_NODE_URL = "https://testnet.mirrornode.hedera.com:443"

// RENDERHIVE CONSTANTS
// #############################################################################
// Account ID of the Renderhive smart contract
const RENDERHIVE_TESTNET_SMART_CONTRACT = "0.0.390082"

// Account IDs of the Hedera Consensus Service topics
// Hive cycle
const RENDERHIVE_TESTNET_TOPIC_HIVE_CYCLE_SYNCHRONIZATION = "0.0.390083"
const RENDERHIVE_TESTNET_TOPIC_HIVE_CYCLE_APPLICATION = "0.0.390084"
const RENDERHIVE_TESTNET_TOPIC_HIVE_CYCLE_VALIDATION = "0.0.390085"

// Render jobs
const RENDERHIVE_TESTNET_RENDER_JOB_QUEUE = "0.0.390086"

// Hive cycle synchronization interval
const RENDERHIVE_CONFIG_HIVE_CYCLE_SYNCHRONIZATION_INTERVAL = 1 * time.Hour

// path to application data
const RENDERHIVE_APP_DIRECTORY_DATA = "data/"
const RENDERHIVE_APP_DIRECTORY_CONFIG = "config/"

// path to local IPFS repository
const RENDERHIVE_APP_DIRECTORY_IPFS_REPO = "ipfs/repo/"

// path to Blender benchmark results
const RENDERHIVE_APP_DIRECTORY_BLENDER_BENCHMARKS = "data/blender_benchmarks/"

// path to the render request documents (both own and from the hive) on this node
const RENDERHIVE_APP_DIRECTORY_LOCAL_REQUESTS = "data/render_requests/local/"
const RENDERHIVE_APP_DIRECTORY_NETWORK_REQUESTS = "data/render_requests/network/"

// BLENDER CONSTANTS
// #############################################################################
// Supported render engines
const (

	// skip the 0 value (reserved for error)
	_ uint8 = iota

	// supported engines
	BLENDER_RENDER_ENGINE_EEVEE
	BLENDER_RENDER_ENGINE_CYCLES

	// all engine options
	BLENDER_RENDER_ENGINE_OPTIONS
)

func GetBlenderEngineString(enum []uint8) []string {
	var result []string

	for _, e := range enum {
		switch e {
		case BLENDER_RENDER_ENGINE_EEVEE:
			result = append(result, "EEVEE")
		case BLENDER_RENDER_ENGINE_CYCLES:
			result = append(result, "CYCLES")

		case BLENDER_RENDER_ENGINE_OPTIONS:
			return []string{"EEVEE", "CYCLES"}

		default:
			fmt.Println(fmt.Errorf("Engine '%v' not in enumeration.", e))
		}
	}
	return result

}

func GetBlenderEngineEnum(engines []string) ([]uint8, error) {
	var result []uint8

	for _, e := range engines {
		switch e {
		case "EEVEE":
			result = append(result, BLENDER_RENDER_ENGINE_EEVEE)
		case "CYCLES":
			result = append(result, BLENDER_RENDER_ENGINE_CYCLES)
		default:
			return []uint8{0}, errors.New(fmt.Sprintf("Invalid engine name '%v'.", e))
		}
	}
	return result, nil
}

// Supported render devices
const (

	// skip the 0 value (reserved for error)
	_ uint8 = iota

	// pure device modes
	BLENDER_RENDER_DEVICE_CPU
	BLENDER_RENDER_DEVICE_CUDA
	BLENDER_RENDER_DEVICE_OPTIX
	BLENDER_RENDER_DEVICE_HIP
	BLENDER_RENDER_DEVICE_ONEAPI
	BLENDER_RENDER_DEVICE_METAL

	// hybrid device modes
	BLENDER_RENDER_DEVICE_CUDA_CPU
	BLENDER_RENDER_DEVICE_OPTIX_CPU
	BLENDER_RENDER_DEVICE_HIP_CPU
	BLENDER_RENDER_DEVICE_ONEAPI_CPU
	BLENDER_RENDER_DEVICE_METAL_CPU

	// all device options
	BLENDER_RENDER_DEVICE_OPTIONS
)

func GetBlenderDeviceString(enum []uint8) []string {
	var result []string

	for _, e := range enum {
		switch e {
		case BLENDER_RENDER_DEVICE_CPU:
			result = append(result, "CPU")
		case BLENDER_RENDER_DEVICE_CUDA:
			result = append(result, "CUDA")
		case BLENDER_RENDER_DEVICE_OPTIX:
			result = append(result, "OPTIX")
		case BLENDER_RENDER_DEVICE_HIP:
			result = append(result, "HIP")
		case BLENDER_RENDER_DEVICE_ONEAPI:
			result = append(result, "ONEAPI")
		case BLENDER_RENDER_DEVICE_METAL:
			result = append(result, "METAL")
		case BLENDER_RENDER_DEVICE_CUDA_CPU:
			result = append(result, "CUDA+CPU")
		case BLENDER_RENDER_DEVICE_OPTIX_CPU:
			result = append(result, "OPTIX+CPU")
		case BLENDER_RENDER_DEVICE_HIP_CPU:
			result = append(result, "HIP+CPU")
		case BLENDER_RENDER_DEVICE_ONEAPI_CPU:
			result = append(result, "ONEAPI+CPU")
		case BLENDER_RENDER_DEVICE_METAL_CPU:
			result = append(result, "METAL+CPU")
		case BLENDER_RENDER_DEVICE_OPTIONS:
			return []string{"CPU", "CUDA", "OPTIX", "HIP", "ONEAPI", "METAL"}

		default:
			fmt.Println(fmt.Errorf("Device '%v' not in enumeration.", e))
		}
	}
	return result

}

func GetBlenderDeviceEnum(devices []string) ([]uint8, error) {
	var result []uint8

	for _, d := range devices {
		switch d {
		case "CPU":
			result = append(result, BLENDER_RENDER_DEVICE_CPU)
		case "CUDA":
			result = append(result, BLENDER_RENDER_DEVICE_CUDA)
		case "OPTIX":
			result = append(result, BLENDER_RENDER_DEVICE_OPTIX)
		case "HIP":
			result = append(result, BLENDER_RENDER_DEVICE_HIP)
		case "ONEAPI":
			result = append(result, BLENDER_RENDER_DEVICE_ONEAPI)
		case "METAL":
			result = append(result, BLENDER_RENDER_DEVICE_METAL)
		case "CUDA+CPU":
			result = append(result, BLENDER_RENDER_DEVICE_CUDA_CPU)
		case "OPTIX+CPU":
			result = append(result, BLENDER_RENDER_DEVICE_OPTIX_CPU)
		case "HIP+CPU":
			result = append(result, BLENDER_RENDER_DEVICE_HIP_CPU)
		case "ONEAPI+CPU":
			result = append(result, BLENDER_RENDER_DEVICE_ONEAPI_CPU)
		case "METAL+CPU":
			result = append(result, BLENDER_RENDER_DEVICE_METAL_CPU)
		default:
			return []uint8{0}, errors.New(fmt.Sprintf("Invalid engine name '%v'.", d))
		}
	}
	return result, nil
}
