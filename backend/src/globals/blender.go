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

// BLENDER VERSION ARCHIVE
// #############################################################################

// DID of the w3up space storing the Blender binaries
const RENDERHIVE_BLENDER_ARCHIVE_DID string = "did:key:z6MkrL1hN8qRwNsynXSvTL14gqQ6ZYg6wwB1zfr1yKr7wbKL"

// specifies the structure of the Blender archive file
type BlenderArchiveFile struct {
	CID      string
	SHA      string
	Commit   string
	Filename string
}
type BlenderArchiveVersion struct {
	Linux   BlenderArchiveFile
	Windows BlenderArchiveFile
	Macos   BlenderArchiveFile
}

// define the version of binaries of Blender stored on the IPFS
var RENDERHIVE_BLENDER_ARCHIVE_FILES = map[string]BlenderArchiveVersion{
	// TODO: Extend the list by more versions
	"4.0.2": {
		Linux: BlenderArchiveFile{
			CID:      "bafybeibetm3w53cyl4kzdopwlad7u7tregokdqpdcdu67xa5kimi4ngxeq",
			SHA:      "17165D5C0A26F263E35C2FD7B5E614BC36898BF9F944A7B0D85D0465B58C9A2F",
			Commit:   "9be62e85b727",
			Filename: "blender-4.0.2-stable+v40.9be62e85b727-linux.x86_64-release.tar.xz",
		},
		Windows: BlenderArchiveFile{},
		Macos:   BlenderArchiveFile{},
	},
}
