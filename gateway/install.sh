#!/bin/sh

type docker &> /dev/null || { echo "Docker is not instaled"; exit 1; }

if !docker info &> /dev/null; then
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
       -v /var/run/docker.sock:/var/run/docker.sock \
       -v /etc/passwd:/etc/passwd \
       -v /etc/group:/etc/group \
       -e SERVER_ADDRESS={{scheme}}://{{host}} \
       -e PRIVATE_KEY=/host/etc/shellhub.key \
       -e TENANT_ID={{tenant_id}} \
       shellhubio/agent:{{version}}
