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

# This Dockerfile describes the containerization of the Renderhive Service App's frontend

# Build the React application
FROM node:18-alpine as builder

# Set the working directory inside the container
WORKDIR /app

# Create a non-root user 'renderhive' and a usergroup 'docker' for the container
RUN addgroup -S docker && adduser -S renderhive -G docker

# Install dependencies and build static assets
COPY frontend/package.json frontend/package-lock.json ./
RUN npm install
COPY frontend/ ./
RUN npm run build

# Serve the React application using Nginx
FROM nginx:alpine

# Copy the build static assets from the builder stage
COPY --from=builder /app/build /usr/share/nginx/html

# Assign ownership of the /app directory to the local container user
RUN chown -R renderhive:docker /app

# Switch to user 'renderhive'
USER renderhive

# Expose port 80
EXPOSE 80

# Start Nginx
CMD ["nginx", "-g", "daemon off;"]