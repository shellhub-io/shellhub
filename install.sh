#!/bin/sh

docker_install() {
    KEEPALIVE_INTERVAL_ARG="-e SHELLHUB_KEEPALIVE_INTERVAL=$KEEPALIVE_INTERVAL"
    PREFERRED_HOSTNAME_ARG="-e SHELLHUB_PREFERRED_HOSTNAME=$PREFERRED_HOSTNAME"
    PREFERRED_IDENTITY_ARG="-e SHELLHUB_PREFERRED_IDENTITY=$PREFERRED_IDENTITY"

    docker run -d \
       --name=$CONTAINER_NAME \
       --restart=on-failure \
       --privileged \
       --net=host \
       --pid=host \
       -v /:/host \
       -v /dev:/dev \
       -v /var/run/docker.sock:/var/run/docker.sock \
       -v /etc/passwd:/etc/passwd \
       -v /etc/group:/etc/group \
       -v /etc/resolv.conf:/etc/resolv.conf \
       -v /var/run:/var/run \
       -v /var/log:/var/log \
       -e SHELLHUB_SERVER_ADDRESS=$SERVER_ADDRESS \
       -e SHELLHUB_PRIVATE_KEY=/host/etc/shellhub.key \
       -e SHELLHUB_TENANT_ID=$TENANT_ID \
       $KEEPALIVE_INTERVAL_ARG \
       $PREFERRED_HOSTNAME_ARG \
       $PREFERRED_IDENTITY_ARG \
       shellhubio/agent:$AGENT_VERSION
}

bundle_install() {
    INSTALL_DIR="${INSTALL_DIR:-/opt/shellhub}"

    if [ $UID -ne 0 ]; then
        echo -e "\nNOTE: This install method requires root privileges\n"
        SUDO="sudo"
    fi

    if ! systemctl show-environment &> /dev/null ; then
        echo "ERROR: This is not a systemd based OS. Could be not proceed.."
        exit 1
    fi


    echo "Downloading runc static binary..."
    {
        download "https://github.com/opencontainers/runc/releases/download/${RUNC_VERSION}/runc.${RUNC_ARCH}" $TMP_DIR/runc && chmod 755 $TMP_DIR/runc
    } || { rm -rf $TMP_DIR && echo "Failed to download runc bnary" && exit 1; }

    echo "Downloading OCI runtime spec file..."
    {
        download "https://raw.githubusercontent.com/shellhub-io/agent/master/config.json" $TMP_DIR/config.json
    } ||  { rm -rf $TMP_DIR && echo "Failed to download OCI runtime spec" && exit 1; }

    echo "Downloading systemd service file..."
    {
        download https://raw.githubusercontent.com/shellhub-io/shellhub/${AGENT_VERSION}/agent/packaging/shellhub.service $TMP_DIR/shellhub.service
    } || { rm -rf $TMP_DIR && echo "Failed to download systemd service file..." && exit 1; }

    echo "Downloading rootfs tarball..."
    {
        download https://github.com/shellhub-io/shellhub/releases/download/$AGENT_VERSION/rootfs-$AGENT_ARCH.tar.gz $TMP_DIR/rootfs.tar.gz
    } || { rm -rf $TMP_DIR && echo "Failed to download rootfs" && exit 1; }

    echo "Extracting rootfs..."
    {
        mkdir -p $TMP_DIR/rootfs && tar -C $TMP_DIR/rootfs -xzf $TMP_DIR/rootfs.tar.gz && rm -f $TMP_DIR/rootfs.tar.gz
    } || { rm -rf $TMP_DIR && echo "Failed to extract rootfs" && exit 1; }

    rm -f $TMP_DIR/rootfs/.dockerenv

    sed -i "s,__SERVER_ADDRESS__,$SERVER_ADDRESS,g" $TMP_DIR/config.json
    sed -i "s,__TENANT_ID__,$TENANT_ID,g" $TMP_DIR/config.json
    sed -i "s,__ROOT_PATH__,$INSTALL_DIR/rootfs,g" $TMP_DIR/config.json
    sed -i "s,__INSTALL_DIR__,$INSTALL_DIR,g" $TMP_DIR/shellhub.service

    $SUDO rm -rf $INSTALL_DIR
    $SUDO mv $TMP_DIR $INSTALL_DIR

    echo "Creating systemd service and starting it"

    $SUDO cp $TMP_DIR/shellhub.service /etc/systemd/system/shellhub.service
    $SUDO systemctl enable --now shellhub-agent || { rm -rf $TMP_DIR && echo "Failed to active systemd service service"; exit 1; }

    rm -rf $TMP_DIR
}

download() {
    local URL=$1
    local OUTPUT=$2

    if type curl > /dev/null 2>&1; then
        curl -fsSL $URL --output $OUTPUT
    elif type wget > /dev/null 2>&1; then
        wget -q -O $OUTPUT $URL
    fi
}

http_get() {
    local URL=$1

    if type curl > /dev/null 2>&1; then
        curl -sk $URL
    elif type wget > /dev/null 2>&1; then
        wget -q -O - $URL
    fi
}

[ -z "$TENANT_ID" ] && { echo "ERROR: TENANT_ID is missing"; exit 1; }

SERVER_ADDRESS="${SERVER_ADDRESS:-https://cloud.shellhub.io}"
TENANT_ID="${TENANT_ID}"
INSTALL_METHOD="$INSTALL_METHOD"
AGENT_VERSION="${AGENT_VERSION:-$(http_get $SERVER_ADDRESS/info | sed -E 's/.*"version":\s?"?([^,"]*)"?.*/\1/')}"
AGENT_ARCH="$AGENT_ARCH"
CONTAINER_NAME="${CONTAINER_NAME:-shellhub}"
RUNC_VERSION=${RUNC_VERSION:-v1.1.3}
RUNC_ARCH=$RUNC_ARCH
INSTALL_DIR="${INSTALL_DIR:-/opt/shellhub}"
TMP_DIR="${TMP_DIR:-`mktemp -d -t shellhub-installer-XXXXXX`}"

if type docker > /dev/null 2>&1; then
    while :; do
        if $SUDO docker infoa > /dev/null 2>&1; then
            INSTALL_METHOD="${INSTALL_METHOD:-docker}"
            break
         elif [ $UID -ne 0 ]; then
            [ -z "$SUDO" ] && SUDO="sudo" || { SUDO="" && break; }
        fi
    done
fi

INSTALL_METHOD="${INSTALL_METHOD:-bundle}"

# Auto detect arch if it has not already been set
if [ -z "$AGENT_ARCH" ]; then
    case `uname -m` in
        x86_64)
            AGENT_ARCH=amd64
            RUNC_ARCH=amd64
            ;;
        armv6l)
            AGENT_ARG=arm32v6
            RUNC_ARHC=armel
            ;;
        armv7l)
            AGENT_ARG=arm32v7
            RUNC_ARHC=armhf
            ;;
        aarch64)
            AGENT_ARG=arm64v8
            RUNC_ARH=arm64
    esac
fi

echo "Install method: $INSTALL_METHOD"
echo "Agent version: $AGENT_VERSION"

case "$INSTALL_METHOD" in
    bundle)
        bundle_install
        ;;
    docker)
        docker_install
        ;;
    *)
        echo "Install method not supported"
        exit 1
esac
