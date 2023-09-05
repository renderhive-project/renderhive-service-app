# ************************** BEGIN LICENSE BLOCK ******************************
#
# Copyright © 2023 Christian Stolze
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
# ************************** END LICENSE BLOCK ********************************

# This Dockerfile describes the containerization of the Renderhive Service App's backend

# COMPILE THE APP
# #############################################################################
# Builder stage to build the Go binary
FROM golang:1.18-alpine as builder

# Set the working directory inside the container
WORKDIR /app/src/

# Copy source code and build the binary
COPY backend/src/ ./
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o renderhive-service ./

# BUILD THE DOCKER IMAGE
# #############################################################################

# Final stage to create a small image
FROM alpine:latest

# install the w3up CLI with the Node.js and npm dependency
RUN apk add --update nodejs npm
RUN npm install -g @web3-storage/w3cli

# Set the working directory inside the container
WORKDIR /app/

# Copy the binary from the builder stage
COPY --from=builder /app/src/renderhive-service /app/renderhive-service

# create required directories
RUN mkdir hedera
RUN mkdir tmp

# Copy the testnet.env into the container
COPY /backend/src/hedera/testnet.env ./hedera/testnet.env

# Run the binary
CMD ["/app/renderhive-service"]