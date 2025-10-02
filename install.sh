#!/bin/sh

# Overridden variables from Go template: {{.Overrides}}
podman_install() {
  [ -n "${KEEPALIVE_INTERVAL}" ] && ARGS="$ARGS -e SHELLHUB_KEEPALIVE_INTERVAL=$KEEPALIVE_INTERVAL"
  [ -n "${PREFERRED_HOSTNAME}" ] && ARGS="$ARGS -e SHELLHUB_PREFERRED_HOSTNAME=$PREFERRED_HOSTNAME"
  [ -n "${PREFERRED_IDENTITY}" ] && ARGS="$ARGS -e SHELLHUB_PREFERRED_IDENTITY=$PREFERRED_IDENTITY"

  echo "📥 Downloading ShellHub container image..."

  {
    $SUDO podman pull -q docker.io/shellhubio/agent:$AGENT_VERSION
  } || {
    echo "❌ Failed to download shellhub container image."
    exit 1
  }

  MODE=""
  DEFAULT_CONTAINER_NAME="shellhub"

  case "$1" in
  "") ;;
  "agent")
    shift 1
    ;;
  "connector")
    MODE="connector"
    DEFAULT_CONTAINER_NAME="shellhub-connector"
    ARGS="$ARGS -e SHELLHUB_PRIVATE_KEYS=${PRIVATE_KEYS:-/host/etc/shellhub/connector/keys}"
    ARGS="$ARGS -e SHELLHUB_CONNECTOR_LABEL=${CONNECTOR_LABEL}"

    echo "🚀 Starting ShellHub container in Docker Connector mode..."
    shift 1
    ;;
  *)
    echo "❌ Invalid mode: $2"
    exit 1
    ;;
  esac

  if [ -z "$MODE" ]; then
    ARGS="$ARGS -e SHELLHUB_PRIVATE_KEY=${PRIVATE_KEY:-/host/etc/shellhub.key}"

    echo "🚀 Starting ShellHub container in Agent mode..."
  fi

  CONTAINER_NAME="${CONTAINER_NAME:-$DEFAULT_CONTAINER_NAME}"

  $SUDO podman run -d \
    --name=$CONTAINER_NAME \
    --replace \
    --restart=on-failure \
    --privileged \
    --pid=host \
    --security-opt label=disable \
    --network host \
    -v /:/host \
    -v /dev:/dev \
    -v /var/run/podman/podman.sock:/var/run/docker.sock \
    -v /proc:/proc \
    -v /var/run:/var/run \
    -v /var/log:/var/log \
    -v /tmp:/tmp \
    -e SHELLHUB_SERVER_ADDRESS=$SERVER_ADDRESS \
    -e SHELLHUB_TENANT_ID=$TENANT_ID \
    $ARGS \
    docker.io/shellhubio/agent:$AGENT_VERSION \
    $MODE
}

docker_install() {
  [ -n "${KEEPALIVE_INTERVAL}" ] && ARGS="$ARGS -e SHELLHUB_KEEPALIVE_INTERVAL=$KEEPALIVE_INTERVAL"
  [ -n "${PREFERRED_HOSTNAME}" ] && ARGS="$ARGS -e SHELLHUB_PREFERRED_HOSTNAME=$PREFERRED_HOSTNAME"
  [ -n "${PREFERRED_IDENTITY}" ] && ARGS="$ARGS -e SHELLHUB_PREFERRED_IDENTITY=$PREFERRED_IDENTITY"

  echo "📥 Downloading ShellHub container image..."

  {
    docker pull -q shellhubio/agent:$AGENT_VERSION
  } || {
    echo "❌ Failed to download shellhub container image."
    exit 1
  }

  MODE=""
  DEFAULT_CONTAINER_NAME="shellhub"

  case "$1" in
  "") ;;
  "agent")
    shift 1
    ;;
  "connector")
    MODE="connector"
    DEFAULT_CONTAINER_NAME="shellhub-connector"
    ARGS="$ARGS -e SHELLHUB_PRIVATE_KEYS=${PRIVATE_KEYS:-/host/etc/shellhub/connector/keys}"
    ARGS="$ARGS -e SHELLHUB_CONNECTOR_LABEL=${CONNECTOR_LABEL}"

    echo "🚀 Starting ShellHub container in Docker Connector mode..."
    shift 1
    ;;
  *)
    echo "❌ Invalid mode: $2"
    exit 1
    ;;
  esac

  if [ -z "$MODE" ]; then
    ARGS="$ARGS -e SHELLHUB_PRIVATE_KEY=${PRIVATE_KEY:-/host/etc/shellhub.key}"

    echo "🚀 Starting ShellHub container in Agent mode..."
  fi

  CONTAINER_NAME="${CONTAINER_NAME:-$DEFAULT_CONTAINER_NAME}"

  $SUDO docker run -d \
    --name=$CONTAINER_NAME \
    --restart=on-failure \
    --privileged \
    --net=host \
    --pid=host \
    -v /:/host \
    -v /dev:/dev \
    -v /var/run/docker.sock:/var/run/docker.sock \
    -v /etc/resolv.conf:/etc/resolv.conf \
    -v /var/run:/var/run \
    -v /var/log:/var/log \
    -v /tmp:/tmp \
    -e SHELLHUB_SERVER_ADDRESS=$SERVER_ADDRESS \
    -e SHELLHUB_TENANT_ID=$TENANT_ID \
    $ARGS \
    shellhubio/agent:$AGENT_VERSION \
    $MODE
}

snap_install() {
  if ! type snap >/dev/null 2>&1; then
    echo "❌ Snap is not installed or not supported on this system."
    exit 1
  fi

  echo "📥 Downloading ShellHub snap package..."

  {
    sudo snap install --classic --channel=latest/stable shellhub
  } || {
    echo "❌ Failed to download and install ShellHub snap package."
    exit 1
  }

  echo "🚀 Starting ShellHub snap service..."

  {
    if [ -n "${PREFERRED_HOSTNAME}" ]; then
      sudo snap set shellhub preferred-hostname="${PREFERRED_HOSTNAME}"
    fi

    sudo snap set shellhub server-address="$SERVER_ADDRESS"
    sudo snap set shellhub tenant-id="$TENANT_ID"
    sudo snap set shellhub private-key="${PRIVATE_KEY:-/etc/shellhub.key}"

    sudo snap start shellhub
  } || {
    echo "❌ Failed to start ShellHub snap service."
    exit 1
  }
}

standalone_install() {
  INSTALL_DIR="${INSTALL_DIR:-/opt/shellhub}"

  if [ "$(id -u)" -ne 0 ]; then
    printf "⚠️ NOTE: This install method requires root privileges.\n"
    SUDO="sudo"
  fi

  if ! systemctl show-environment >/dev/null 2>&1; then
    printf "❌ ERROR: This is not a systemd-based operation system. Unable to proceed with the requested action.\n"
    exit 1
  fi

  echo "📥 Downloading required files..."

  {
    download "https://github.com/opencontainers/runc/releases/download/${RUNC_VERSION}/runc.${RUNC_ARCH}" $TMP_DIR/runc && chmod 755 $TMP_DIR/runc
  } || { rm -rf $TMP_DIR && echo "❌ Failed to download runc binary." && exit 1; }

  {
    download https://raw.githubusercontent.com/shellhub-io/shellhub/${AGENT_VERSION}/agent/packaging/config.json $TMP_DIR/config.json
  } || { rm -rf $TMP_DIR && echo "❌ Failed to download OCI runtime spec." && exit 1; }

  {
    download https://raw.githubusercontent.com/shellhub-io/shellhub/${AGENT_VERSION}/agent/packaging/shellhub-agent.service $TMP_DIR/shellhub-agent.service
  } || { rm -rf $TMP_DIR && echo "❌ Failed to download systemd service file." && exit 1; }

  {
    download https://github.com/shellhub-io/shellhub/releases/download/$AGENT_VERSION/rootfs-$AGENT_ARCH.tar.gz $TMP_DIR/rootfs.tar.gz
  } || { rm -rf $TMP_DIR && echo "❌ Failed to download rootfs." && exit 1; }

  echo "📂 Extracting files..."

  {
    mkdir -p $TMP_DIR/rootfs && tar -C $TMP_DIR/rootfs -xzf $TMP_DIR/rootfs.tar.gz && rm -f $TMP_DIR/rootfs.tar.gz
  } || { rm -rf $TMP_DIR && echo "❌ Failed to extract rootfs." && exit 1; }

  rm -f $TMP_DIR/rootfs/.dockerenv

  sed -i "s,__SERVER_ADDRESS__,$SERVER_ADDRESS,g" $TMP_DIR/config.json
  sed -i "s,__TENANT_ID__,$TENANT_ID,g" $TMP_DIR/config.json
  sed -i "s,__ROOT_PATH__,$INSTALL_DIR/rootfs,g" $TMP_DIR/config.json
  sed -i "s,__INSTALL_DIR__,$INSTALL_DIR,g" $TMP_DIR/shellhub-agent.service

  echo "🚀 Starting ShellHub system service..."

  $SUDO cp $TMP_DIR/shellhub-agent.service /etc/systemd/system/shellhub-agent.service

  # NOTE: As we need to check if the service is running to indicate it was installed correctly, we need to copy the
  # values to install directory before enable it, to a correctly check the status.
  $SUDO rm -rf $INSTALL_DIR
  $SUDO mv $TMP_DIR $INSTALL_DIR

  uninstall() {
    echo "Please check the logs with the command:"
    echo "journalctl -f -u shellhub-agent"
    echo ""
    echo "❗ Uninstalling ShellHub agent..."
    $SUDO rm -rf $TMP_DIR
    $SUDO rm -rf $INSTALL_DIR
    $SUDO rm /etc/systemd/system/shellhub-agent.service
  }

  $SUDO systemctl enable --now shellhub-agent || {
    uninstall && echo "❌ Failed to enable systemd service."
    exit 1
  }

  trap 'echo "❗ Interrupted. Disabling shellhub-agent..."; $SUDO systemctl disable --now shellhub-agent; exit 1' INT

  echo "🔍 Checking service status..."
  echo "Please wait for the service to start. This may take a few seconds."
  echo "Press Ctrl+C to cancel the installation."

  timeout 15s sh -c '
      journalctl -f -u shellhub-agent --since "$(systemctl show -p ActiveEnterTimestamp shellhub-agent | cut -d= -f2)" | while read -r line; do
        if echo "$line" | grep -Eq "Listening for connections"; then
            echo "✅ Success: $line"
            exit 0
        elif echo "$line" | grep -Eq "fatal"; then
            echo "❌ Failure: $line"
            exit 2
        fi
      done
    '

  exit_code=$?

  if [ $exit_code -eq 124 ]; then
    echo "❌ Timeout: Service took too long to start."
    echo "Disabling shellhub-agent service..."
    $SUDO systemctl disable --now shellhub-agent
    uninstall
    exit 1
  elif [ $exit_code -eq 2 ]; then
    echo "Disabling shellhub-agent service..."
    $SUDO systemctl disable --now shellhub-agent
    uninstall
    exit 1
  fi

  $SUDO rm -rf $TMP_DIR
}

wsl_install() {
  if ! systemctl show-environment >/dev/null 2>&1; then
    printf "❌ ERROR: This install method requires systemd to be enabled.\n"
    printf "Please refer to the following link for instructions on how to enable systemd:\n"
    printf "https://learn.microsoft.com/en-us/windows/wsl/wsl-config#systemd-support\n"
    printf "Once systemd is enabled, run this script again to complete the installation.\n"
    exit 1
  fi

  if [ "$(wslinfo --networking-mode)" != "mirrored" ]; then
    printf "❌ ERROR: WSL networking mode must be set to mirrored.\n"
    printf "Please refer to the following link for instructions on how to set the networking mode:\n"
    printf "https://learn.microsoft.com/en-us/windows/wsl/networking#mirrored-mode-networking\n"
    printf "Once the networking mode is set to mirrored, run this script again to complete the installation.\n"
    exit 1
  fi

  standalone_install
}

download() {
  _DOWNLOAD_URL=$1
  _DOWNLOAD_OUTPUT=$2

  if type curl >/dev/null 2>&1; then
    curl -fsSL $_DOWNLOAD_URL --output $_DOWNLOAD_OUTPUT
  elif type wget >/dev/null 2>&1; then
    wget -q -O $_DOWNLOAD_OUTPUT $_DOWNLOAD_URL
  fi
}

http_get() {
  _HTTP_GET_URL=$1

  if type curl >/dev/null 2>&1; then
    curl -sk $_HTTP_GET_URL
  elif type wget >/dev/null 2>&1; then
    wget -q -O - $_HTTP_GET_URL
  fi
}

if [ "$(uname -s)" = "FreeBSD" ]; then
  echo "👹 This system is running FreeBSD."
  echo "❌ ERROR: Automatic installation is not supported on FreeBSD."
  echo
  echo "Please refer to the ShellHub port at https://github.com/shellhub-io/ports"
  exit 1
fi

[ -z "$TENANT_ID" ] && {
  echo "ERROR: TENANT_ID is missing."
  exit 1
}

SERVER_ADDRESS="${SERVER_ADDRESS:-https://cloud.shellhub.io}"
TENANT_ID="${TENANT_ID}"
INSTALL_METHOD="$INSTALL_METHOD"
AGENT_VERSION="${AGENT_VERSION:-$(http_get $SERVER_ADDRESS/info | sed -E 's/.*"version":\s?"?([^,"]*)"?.*/\1/')}"
AGENT_ARCH="$AGENT_ARCH"
RUNC_VERSION=${RUNC_VERSION:-v1.1.3}
RUNC_ARCH=$RUNC_ARCH
INSTALL_DIR="${INSTALL_DIR:-/opt/shellhub}"
TMP_DIR="${TMP_DIR:-$(mktemp -d -t shellhub-installer-XXXXXX)}"

# Auto detect arch if it has not already been set
if [ -z "$AGENT_ARCH" ]; then
  case $(uname -m) in
  x86_64)
    AGENT_ARCH=amd64
    RUNC_ARCH=amd64
    ;;
  armv6l)
    AGENT_ARCH=arm32v6
    RUNC_ARCH=armel
    ;;
  armv7l)
    AGENT_ARCH=arm32v7
    RUNC_ARCH=armhf
    ;;
  aarch64)
    AGENT_ARCH=arm64v8
    RUNC_ARCH=arm64
    ;;
  esac
fi

echo "🛠️ ShellHub Agent Installer"
echo
if [ -z "$INSTALL_METHOD" ]; then
  echo "This script will install the ShellHub agent on your system."
  echo "It will auto-detect the best available installation method."
  echo
  echo "Installation methods (priority order):"
  echo "  1. Docker     - If Docker is installed and accessible in rootful mode"
  echo "  2. Podman     - If Podman is installed and accessible in rootful mode"
  echo "  3. Snap       - If Snap package manager is available"
  echo "  4. WSL        - If running in WSL2 with systemd and mirrored networking"
  echo "  5. Standalone - Fallback method using runc and systemd"
  echo
fi

echo "⚙️ Detected settings:"
echo "- Server address: $SERVER_ADDRESS"
echo "- Tenant ID: $TENANT_ID"
echo "- Agent version: $AGENT_VERSION"
echo "- Architecture: $AGENT_ARCH"
[ -n "$INSTALL_METHOD" ] && echo "- Install method: $INSTALL_METHOD"
echo

if [ -z "$INSTALL_METHOD" ] && type docker >/dev/null 2>&1; then
  echo "🔍 Checking if Docker is available and accessible in rootful mode..."

  export DOCKER_HOST="${DOCKER_HOST:-unix:///var/run/docker.sock}"

  for prefix in "" "sudo"; do
    if $prefix docker info >/dev/null 2>&1; then
      SUDO=$prefix
      INSTALL_METHOD="docker"
      break
    fi
  done

  [ -z "$INSTALL_METHOD" ] && echo "ℹ️ Docker is not accessible in rootful mode."
fi

if [ -z "$INSTALL_METHOD" ] && type podman >/dev/null 2>&1; then
  echo "🔍 Checking if Podman is available and accessible in rootful mode..."

  export CONTAINER_HOST="${CONTAINER_HOST:-unix:///var/run/podman/podman.sock}"

  for prefix in "" "sudo"; do
    if $prefix podman info >/dev/null 2>&1; then
      SUDO=$prefix
      INSTALL_METHOD="podman"
      break
    fi
  done

  [ -z "$INSTALL_METHOD" ] && echo "ℹ️ Podman is not accessible in rootful mode."
fi

if [ -z "$INSTALL_METHOD" ]; then
  echo
  echo "⚠️  NOTE: No recommended installation method was detected."
  echo "⚠️  For best performance, easier updates, and better isolation, it is strongly recommended to use Docker or Podman."
  echo "ℹ️  The installer will proceed with an alternative method (Snap, Standalone, or WSL), but these may have limitations."
  echo
fi

if [ -z "$INSTALL_METHOD" ] && type snap >/dev/null 2>&1; then
  echo "🔍 Detected Snap package manager..."
  INSTALL_METHOD="snap"
fi

# Check if running on WSL
if grep -qi Microsoft /proc/version; then
  echo "🔍 Detected WSL environment..."

  WSL_EXE=$(find /mnt/*/Windows/System32/wsl.exe 2>/dev/null | head -n 1)
  WSL_VERSION=$($WSL_EXE -v | tr -d '\0' | grep "WSL version" | awk -F'[ .:]+' '{print $3}')

  if [ -z "$WSL_VERSION" ] || [ "$WSL_VERSION" -lt 2 ]; then
    echo "❌ ERROR: WSL version 2 is required to run ShellHub."
    exit 1
  fi

  if  grep -qi 'NAME="Ubuntu"' /etc/os-release; then
    INSTALL_METHOD="wsl"
  else
    echo "❌ Error: Only Ubuntu is supported in WSL."
    exit 1
  fi
fi

[ -z "$INSTALL_METHOD" ] && INSTALL_METHOD="standalone"

case "$INSTALL_METHOD" in
podman)
  echo "🐳 Installing ShellHub using podman method..."
  podman_install "$@"
  ;;
docker)
  echo "🐳 Installing ShellHub using docker method..."
  docker_install "$@"
  ;;
snap)
  echo "📦 Installing ShellHub using snap method..."
  snap_install
  ;;
standalone)
  echo "🐧 Installing ShellHub using standalone method..."
  standalone_install
  ;;
wsl)
  echo "🪟 Installing ShellHub using WSL method..."
  wsl_install
  ;;
*)
  echo "❌ Install method not supported."
  exit 1
  ;;
esac
