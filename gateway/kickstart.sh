#!/bin/sh

# This file is a template which gets run through the "/kickstart.sh" route to
# generate the final script file. PLEASE NEVER RUN THIS FILE DIRECTLY, instead
# run me with:
#
# curl -sSf "http://<SERVER-ADDRESS>/kickstart.sh?tenant_id=<TENANT-ID>" | sh
#
# Where:
# <SERVER-ADDRESS> is the ShellHub server address
# <TENANT-ID> is your tenant ID
#
# List of URL parameters of /kickstart.sh URL:
#
# keepalive_interval = Specifies in seconds the keep alive message interval
# preferred_hostname = The preferred hostname to use rather than generated
#                      value from ethernet MAC address

type docker > /dev/null 2>&1 || { echo "Docker is not instaled"; exit 1; }

if ! docker info > /dev/null 2>&1; then
    cat <<EOF
Docker is not running or your current user is not in docker group.
You need to manually add your current user to docker group or run this installer using sudo.

EOF
    while true; do
	echo -n "Do you want to run the installer using sudo? [y/N] "
	read yn
	case $yn in
            [Yy]|YES|yes) SUDO="sudo"; break;;
            [Nn]|NO|no|"") echo "It cannot proceed, exiting"; exit;;
            *) echo "Please answer 'yes' or 'no'.";;
	esac
    done
fi

$SUDO docker run -d \
       --name=shellhub \
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
       -e SHELLHUB_SERVER_ADDRESS={{scheme}}://{{host}} \
       -e SHELLHUB_PRIVATE_KEY=/host/etc/shellhub.key \
       -e SHELLHUB_TENANT_ID={{tenant_id}} \
       {% if keepalive_interval ~= '' and keepalive_interval ~= nil then %}
       -e SHELLHUB_KEEPALIVE_INTERVAL={{keepalive_interval}} \
       {% end %}
       {% if preferred_hostname ~= '' and preferred_hostname ~= nil then %}
       -e SHELLHUB_PREFERRED_HOSTNAME={{preferred_hostname}} \
       {% end %}
       {% if preferred_identity ~= '' and preferred_identity ~= nil then %}
       -e SHELLHUB_PREFERRED_IDENTITY={{preferred_identity}} \
       {% end %}
       shellhubio/agent:{{version}}
