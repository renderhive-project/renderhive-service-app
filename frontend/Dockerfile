# ************************** BEGIN LICENSE BLOCK ******************************
#
# Copyright © 2024 Christian Stolze
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

# This Dockerfile describes the containerization of the Renderhive Service App's frontend

# COMPILE THE FRONTEND SOURCE CODE
# #############################################################################
# Build the React application
FROM node:18-alpine as builder

# Set the working directory inside the container
WORKDIR /app

# Install dependencies and build static assets
COPY frontend/package.json frontend/package-lock.json ./
RUN npm install
COPY frontend/ ./
RUN npm run build


# BUILD THE DOCKER IMAGE
# #############################################################################
# Serve the React application using Nginx
FROM nginx:alpine

# Copy the build static assets from the builder stage
COPY --from=builder /app/dist /usr/share/nginx/html

# Copy SSL/TLS certificates
COPY server/certificates /etc/nginx/ssl

# Copy Nginx configuration file
COPY server/nginx/nginx.conf /etc/nginx/nginx.conf

# Expose ports 80 and 443
EXPOSE 80
EXPOSE 443

# Start Nginx
CMD ["nginx", "-g", "daemon off;"]