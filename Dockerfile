# Minimal CI image for twiggit project with Alpine and basic dependencies
FROM alpine:3.22

# Accept mise version as build argument
ARG MISE_VERSION=v2025.9.13

# Install minimal system dependencies needed for project
RUN apk add --no-cache bash curl git ca-certificates coreutils build-base
RUN update-ca-certificates

# Set working directory
WORKDIR /app

# Copy mise configurations
COPY .mise/ .mise/

# Set CI environment for mise
ENV MISE_ENV=ci 

RUN curl https://mise.run | MISE_VERSION=$MISE_VERSION MISE_INSTALL_PATH=/usr/local/bin/mise sh

# Set bash as default shell and enable pipefail
SHELL ["/bin/bash", "-o", "pipefail", "-c"]

# Label the image
LABEL maintainer="twiggit"
LABEL description="Minimal CI image for twiggit with cache-based tool management"