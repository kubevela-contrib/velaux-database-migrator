#!/usr/bin/env bash

# Implemented based on VelaD Cli https://github.com/kubevela/velad

# Velamg location
: ${VELAMG_INSTALL_DIR:="/usr/local/bin"}

# sudo is required to copy binary to VELAMG_INSTALL_DIR for linux
: ${USE_SUDO:="false"}

# Http request CLI
VELAMG_HTTP_REQUEST_CLI=curl

# Velamg filename
VELAMG_CLI_FILENAME=velaux-database-migrator

VELAMG_NEW_FILENAME=velamg

VELAMG_CLI_FILE="${VELAMG_INSTALL_DIR}/${VELAMG_NEW_FILENAME}"

DOWNLOAD_BASE="https://github.com/kubevela-contrib/velaux-database-migrator/releases/download"
API_BASE="https://api.github.com/repos/kubevela-contrib/velaux-database-migrator/releases"

getSystemInfo() {
    ARCH=$(uname -m)
    case $ARCH in
        armv7*) ARCH="arm";;
        aarch64) ARCH="arm64";;
        x86_64) ARCH="amd64";;
    esac

    OS=$(echo `uname`|tr '[:upper:]' '[:lower:]')

    # Most linux distro needs root permission to copy the file to /usr/local/bin
    if [ "$OS" == "linux" ] || [ "$OS" == "darwin" ]; then
        if [ "$VELAMG_INSTALL_DIR" == "/usr/local/bin" ]; then
            USE_SUDO="true"
        fi
    fi
}

verifySupported() {
    local supported=(darwin-amd64 linux-amd64 linux-arm64 darwin-arm64)
    local current_osarch="${OS}-${ARCH}"

    for osarch in "${supported[@]}"; do
        if [ "$osarch" == "$current_osarch" ]; then
            echo "Your system is ${OS}_${ARCH}"
            return
        fi
    done

    echo "No prebuilt binary for ${current_osarch}"
    exit 1
}

runAsRoot() {
    local CMD="$*"

    if [ $EUID -ne 0 -a $USE_SUDO = "true" ]; then
        CMD="sudo $CMD"
    fi

    $CMD
}

checkHttpRequestCLI() {
    if type "curl" > /dev/null; then
        VELAMG_HTTP_REQUEST_CLI=curl
    elif type "wget" > /dev/null; then
        VELAMG_HTTP_REQUEST_CLI=wget
    else
        echo "Either curl or wget is required"
        exit 1
    fi
}

checkExistingVelamg() {
    if [ -f "$VELAMG_CLI_FILE" ]; then
        echo -e "\nVelamg is detected:"
        echo -e "Reinstalling Velamg - ${VELAMG_CLI_FILE}...\n"
    else
        echo -e "Installing Velamg ...\n"
    fi
}

getLatestRelease() {
    local velamgReleaseUrl="${API_BASE}/latest"
    local latest_release=""

    if [ "$VELAMG_HTTP_REQUEST_CLI" == "curl" ]; then
         latest_release=$(curl --silent $velamgReleaseUrl |
                              grep '"tag_name":' |
                              sed -E 's/.*"([^"]+)".*/\1/' |
                              sed -n 's/v\([0-9.]*\)/\1/p' )
    else
        latest_release=$(wget -q -O - $velamgReleaseUrl |
                                      grep '"tag_name":' |
                                      sed -E 's/.*"([^"]+)".*/\1/' |
                                      sed -n 's/v\([0-9.]*\)/\1/p' )
    fi
    echo $latest_release
    ret_val=$latest_release
}

downloadFile() {
    LATEST_RELEASE_TAG=$1

    VELAMG_CLI_ARTIFACT="${VELAMG_CLI_FILENAME}_${LATEST_RELEASE_TAG}_${OS}_${ARCH}.tar.gz"
    DOWNLOAD_URL="${DOWNLOAD_BASE}/v${LATEST_RELEASE_TAG}/${VELAMG_CLI_ARTIFACT}"

    # Create the temp directory
    VELAMG_TMP_ROOT=$(mktemp -dt velamg-install-XXXXXX)
    ARTIFACT_TMP_FILE="$VELAMG_TMP_ROOT/$VELAMG_CLI_ARTIFACT"

    echo "Downloading $DOWNLOAD_URL ..."
    if [ "$VELAMG_HTTP_REQUEST_CLI" == "curl" ]; then
        curl -SL "$DOWNLOAD_URL" -o "$ARTIFACT_TMP_FILE"
    else
        wget -O "$ARTIFACT_TMP_FILE" "$DOWNLOAD_URL"
    fi

    if [ ! -f "$ARTIFACT_TMP_FILE" ]; then
        echo "failed to download $DOWNLOAD_URL ..."
        exit 1
    fi
}

installFile() {
    tar xf "$ARTIFACT_TMP_FILE" -C "$VELAMG_TMP_ROOT"
    mv "$VELAMG_TMP_ROOT/$VELAMG_CLI_FILENAME" "$VELAMG_TMP_ROOT/$VELAMG_NEW_FILENAME"
    local tmp_root_velamg="$VELAMG_TMP_ROOT/$VELAMG_NEW_FILENAME"

    if [ ! -f "$tmp_root_velamg" ]; then
        echo "Failed to unpack Velamg executable."
        exit 1
    fi

    chmod o+x "$tmp_root_velamg"
    runAsRoot cp "$tmp_root_velamg" "$VELAMG_INSTALL_DIR"

    if [ $? -eq 0 ] && [ -f "$VELAMG_CLI_FILE" ]; then
        echo "Velamg installed into $VELAMG_INSTALL_DIR/$VELAMG_NEW_FILENAME successfully."
        echo ""
        $VELAMG_CLI_FILE help
    else
        echo "Failed to install $VELAMG_NEW_FILENAME"
        exit 1
    fi
}

fail_trap() {
    result=$?
    if [ "$result" != "0" ]; then
        echo "Failed to install Velamg"
        echo "Go to https://github.com/kubevela-contrib/velaux-database-migrator for more support."
    fi
    cleanup
    exit $result
}

cleanup() {
    if [[ -d "${VELAMG_TMP_ROOT:-}" ]]; then
        rm -rf "$VELAMG_TMP_ROOT"
    fi
}

installCompleted() {
    echo -e "\nFor more information on how to started, please visit:"
    echo -e "  https://github.com/kubevela-contrib/velaux-database-migrator"
}

# -----------------------------------------------------------------------------
# main
# -----------------------------------------------------------------------------
trap "fail_trap" EXIT

getSystemInfo
verifySupported
checkExistingVelamg
checkHttpRequestCLI


if [ -z "$1" ]; then
    echo "Getting the latest Velamg..."
    getLatestRelease
elif [[ $1 == v* ]]; then
    ret_val=$1
else
    ret_val=v$1
fi

downloadFile $ret_val
installFile
cleanup

installCompleted