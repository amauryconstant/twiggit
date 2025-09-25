# Minimal CI image for twiggit project with Alpine and basic dependencies
ARG ALPINE_VERSION=3.22
FROM alpine:$ALPINE_VERSION

# Accept versions as build arguments
ARG MISE_VERSION
ARG BASH_VERSION
ARG ZSH_VERSION
ARG FISH_VERSION

# Install minimal system dependencies needed for project
RUN apk add --no-cache bash${BASH_VERSION:+=$BASH_VERSION} zsh${ZSH_VERSION:+=$ZSH_VERSION} fish${FISH_VERSION:+=$FISH_VERSION} curl git ca-certificates coreutils build-base
RUN update-ca-certificates

# Set working directory
WORKDIR /app

# Copy mise configurations
COPY .mise/ .mise/

# Set CI environment for mise
ENV MISE_ENV=ci 

# Install mise
RUN curl https://mise.run | MISE_VERSION=$MISE_VERSION MISE_INSTALL_PATH=/usr/local/bin/mise sh
RUN curl https://mise.run | MISE_VERSION=$MISE_VERSION MISE_INSTALL_PATH=/usr/local/bin/mise sh


# Set bash as default shell and enable pipefail
SHELL ["/bin/bash", "-o", "pipefail", "-c"]

# Label the image
LABEL maintainer="twiggit"
LABEL description="Minimal CI image for twiggit with cache-based tool management"
