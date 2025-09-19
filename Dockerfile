# Custom CI image for twiggit project with Alpine and mise
FROM alpine:3.22

# Install dependencies needed for the project
RUN apk add --no-cache \
        bash \
        curl \
        git \
        tar \
        sudo \
        ca-certificates \
        bash-completion \
        openssl \
        coreutils \
    && update-ca-certificates

# Set environment variables for mise
ENV MISE_VERSION=v2025.9.13 \
    MISE_SHA256=ff15a888170bbdb8d976fc9abe62c5ec96102ba5fb516139b959103664de6439 \
    MISE_DATA_DIR="/mise" \
    MISE_CONFIG_DIR="/mise" \
    MISE_CACHE_DIR="/mise/cache" \
    MISE_INSTALL_PATH="/usr/local/bin/mise" \
    SHELL="/bin/bash"

# Set bash as default shell and enable pipefail
SHELL ["/bin/bash", "-o", "pipefail", "-c"]

# Install mise securely with version pinning and checksum verification
RUN apk add --no-cache --virtual .build-deps wget \
    && wget -O /usr/local/bin/mise "https://github.com/jdx/mise/releases/download/${MISE_VERSION}/mise-${MISE_VERSION}-linux-x64-musl" \
    && echo "${MISE_SHA256}  /usr/local/bin/mise" | sha256sum -c - \
    && chmod +x /usr/local/bin/mise \
    && apk del .build-deps

# Set working directory
WORKDIR /app

# Verify mise installation
RUN mise --version

# Label the image
LABEL maintainer="twiggit"
LABEL description="Custom CI image for twiggit with pre-installed mise on Alpine"
