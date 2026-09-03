#!/usr/bin/env bats
# Tests for install.sh itself: its functions and its install-method detection.
# The wrapper it generates is a separate artifact, covered by
# shellhub-agent-wrapper.bats.

load install.helpers

setup() {
    setup_install_env

    stub_bin id 'echo 0'
    skip_enrollment_wait

    export SERVER_ADDRESS="https://cloud.example.test"
    export AGENT_VERSION="v0.0.0-test"
    export AGENT_IMAGE="shellhubio/agent:v0.0.0-test"
    export INSTALL_DIR="$BATS_TEST_TMPDIR/usr-local-bin"
    export TMP_DIR="$BATS_TEST_TMPDIR/tmp"
    export AGENT_KEY="$BATS_TEST_TMPDIR/shellhub.key"
    export PRIVATE_KEY="/host$AGENT_KEY"
    export PROC_VERSION="$BATS_TEST_TMPDIR/proc-version"
    export OS_RELEASE="$BATS_TEST_TMPDIR/os-release"

    mkdir -p "$INSTALL_DIR" "$TMP_DIR"
    echo "Linux version 6.8.0-generic" > "$PROC_VERSION"
    echo 'NAME="Ubuntu"' > "$OS_RELEASE"

    RUNTIME=docker
}

skip_enrollment_wait() {
    touch "$BATS_TEST_TMPDIR/shellhub.key"
    stub_bin sleep 'exit 0'
}

with_tenant() {
    export TENANT_ID="00000000-0000-4000-a000-000000000000"
}

as_non_root() {
    stub_bin id 'echo 1000'
    stub_sudo
}

stub_sudo() {
    stub_bin sudo 'echo "sudo $*" >> "$CALLS"; export RAN_UNDER_SUDO=1; exec "$@"'
}

use_podman() {
    RUNTIME=podman
    stub_bin systemctl 'exit 0'
}

container_install() {
    [ -x "$STUB_DIR/$RUNTIME" ] || stub_bin "$RUNTIME"
    call_install "${RUNTIME}_install" "$@"
}

fake_agent_binary() {
    export AGENT_BINARY="$BATS_TEST_TMPDIR/fake-agent"
    printf '#!/bin/sh\necho "agent $*" >> "$CALLS"\nexit %s\n' "${1:-0}" > "$AGENT_BINARY"
    chmod +x "$AGENT_BINARY"
}

enter_wsl() {
    echo "Linux version 5.15.0-microsoft-standard-WSL2 Microsoft" > "$PROC_VERSION"
    stub_bin find 'echo "$STUB_DIR/wsl.exe"'
    stub_bin wsl.exe "echo \"WSL version: ${1:-2.0.0.0}\""
}

@test "check_podman_boot_restart aborts when systemctl is missing" {
    call_install check_podman_boot_restart

    [ "$status" -eq 1 ]
    assert_output_contains "systemctl not found"
    assert_output_contains "SKIP_BOOT_RESTART_CHECK=1"
}

@test "check_podman_boot_restart aborts when podman-restart.service is disabled" {
    stub_bin systemctl 'echo "systemctl $*" >> "$CALLS"; exit 1'

    call_install check_podman_boot_restart

    [ "$status" -eq 1 ]
    assert_called "systemctl is-enabled podman-restart.service"
    assert_output_contains "podman-restart.service is not enabled"
    assert_output_contains "sudo systemctl enable podman-restart.service"
}

@test "check_podman_boot_restart passes silently when podman-restart.service is enabled" {
    stub_bin systemctl 'echo "systemctl $*" >> "$CALLS"; exit 0'

    call_install check_podman_boot_restart

    [ "$status" -eq 0 ]
    [ -z "$output" ]
    assert_called "systemctl is-enabled podman-restart.service"
}

@test "SKIP_BOOT_RESTART_CHECK installs anyway when systemctl is missing" {
    export SKIP_BOOT_RESTART_CHECK=1

    call_install check_podman_boot_restart

    [ "$status" -eq 0 ]
    assert_output_contains "skipping podman-restart.service check"
    assert_output_contains "will NOT auto-start on boot"
}

@test "SKIP_BOOT_RESTART_CHECK installs anyway when podman-restart.service is disabled" {
    export SKIP_BOOT_RESTART_CHECK=1
    stub_bin systemctl 'echo "systemctl $*" >> "$CALLS"; exit 1'

    call_install check_podman_boot_restart

    [ "$status" -eq 0 ]
    refute_called "systemctl is-enabled"
    assert_output_contains "skipping podman-restart.service check"
}

@test "enrollment_summary reports a pairing code ahead of every other credential" {
    with_tenant
    export CODE=ABC123 INSTALL_KEY=key-1

    call_install enrollment_summary

    [ "$output" = "pairing code (pre-authorized)" ]
}

@test "enrollment_summary reports an install key ahead of a tenant" {
    with_tenant
    export INSTALL_KEY=key-1

    call_install enrollment_summary

    [ "$output" = "install key" ]
}

@test "enrollment_summary reports a tenant as landing pending" {
    with_tenant

    call_install enrollment_summary

    [ "$output" = "tenant $TENANT_ID (device lands pending)" ]
}

@test "enrollment_summary points at the login flow when no credential is given" {
    call_install enrollment_summary

    [ "$output" = "none — enroll with 'shellhub-agent login'" ]
}

@test "enroll_agent_interactively runs the login flow when no credential names a namespace" {
    stub_bin shellhub-agent

    call_install enroll_agent_interactively shellhub-agent "$AGENT_KEY"

    [ "$status" -eq 0 ]
    assert_called "shellhub-agent login"
}

@test "enroll_agent_interactively waits for the agent key before starting the login flow" {
    rm -f "$AGENT_KEY"
    stub_bin sleep 'echo "sleep $*" >> "$CALLS"'
    stub_bin shellhub-agent

    call_install enroll_agent_interactively shellhub-agent "$AGENT_KEY"

    [ "$status" -eq 0 ]
    [ "$(grep -c '^sleep 1$' "$CALLS")" -eq 30 ]
    assert_called "shellhub-agent login"
}

@test "enroll_agent_interactively does not wait when the agent key is already on disk" {
    stub_bin sleep 'echo "sleep $*" >> "$CALLS"'
    stub_bin shellhub-agent

    call_install enroll_agent_interactively shellhub-agent "$AGENT_KEY"

    [ "$status" -eq 0 ]
    refute_called "sleep"
}

@test "enroll_agent_interactively survives a login the user abandons" {
    stub_bin shellhub-agent 'echo "shellhub-agent $*" >> "$CALLS"; exit 130'

    call_install enroll_agent_interactively shellhub-agent "$AGENT_KEY"

    [ "$status" -eq 0 ]
}

@test "enroll_agent_interactively skips the login flow for a pairing code" {
    export CODE=ABC123
    stub_bin shellhub-agent

    call_install enroll_agent_interactively shellhub-agent "$AGENT_KEY"

    [ "$status" -eq 0 ]
    refute_called "shellhub-agent"
    assert_output_contains "pre-authorized"
}

@test "enroll_agent_interactively skips the login flow for an install key" {
    export INSTALL_KEY=key-1
    stub_bin shellhub-agent

    call_install enroll_agent_interactively shellhub-agent "$AGENT_KEY"

    [ "$status" -eq 0 ]
    refute_called "shellhub-agent"
    assert_output_contains "install key's namespace"
}

@test "enroll_agent_interactively skips the login flow for a tenant" {
    with_tenant
    stub_bin shellhub-agent

    call_install enroll_agent_interactively shellhub-agent "$AGENT_KEY"

    [ "$status" -eq 0 ]
    refute_called "shellhub-agent"
    assert_output_contains "appear as pending in the console"
}

@test "docker_install runs the container with unless-stopped so it survives a reboot" {
    container_install

    [ "$status" -eq 0 ]
    assert_called "docker run -d --name=shellhub --restart=unless-stopped"
}

@test "podman_install runs the container with unless-stopped so podman-restart.service picks it up" {
    use_podman

    container_install

    [ "$status" -eq 0 ]
    assert_called "podman run -d --name=shellhub --replace --restart=unless-stopped"
}

@test "docker_install grants the agent the host access it needs" {
    container_install

    [ "$status" -eq 0 ]
    assert_called "--privileged --net=host --pid=host"
    assert_called "-v /:/host"
    assert_called "-v /var/run/docker.sock:/var/run/docker.sock"
    assert_called "-v /etc/resolv.conf:/etc/resolv.conf"
}

@test "podman_install grants the agent the host access it needs" {
    use_podman

    container_install

    [ "$status" -eq 0 ]
    assert_called "--privileged --pid=host --security-opt label=disable --network host"
    assert_called "-v /:/host"
}

@test "podman_install presents the podman socket to the agent as the docker socket" {
    use_podman

    container_install

    [ "$status" -eq 0 ]
    assert_called "-v /var/run/podman/podman.sock:/var/run/docker.sock"
}

@test "docker_install labels the agent container so the wrapper can find it" {
    container_install

    [ "$status" -eq 0 ]
    assert_called "--label shellhub.role=agent"
    assert_output_contains "Starting ShellHub container in Agent mode"
}

@test "podman_install labels the agent container so the wrapper can find it" {
    use_podman

    container_install

    [ "$status" -eq 0 ]
    assert_called "--label shellhub.role=agent"
}

@test "docker_install treats an explicit agent mode as the default mode" {
    container_install agent

    [ "$status" -eq 0 ]
    assert_called "--label shellhub.role=agent"
    assert_output_contains "Starting ShellHub container in Agent mode"
}

@test "podman_install treats an explicit agent mode as the default mode" {
    use_podman

    container_install agent

    [ "$status" -eq 0 ]
    assert_called "--label shellhub.role=agent"
}

@test "docker_install omits the agent label in connector mode" {
    with_tenant

    container_install connector

    [ "$status" -eq 0 ]
    refute_called "--label shellhub.role=agent"
    assert_called "--name=shellhub-connector"
    assert_called "-e SHELLHUB_PRIVATE_KEYS=/host/etc/shellhub/connector/keys"
}

@test "podman_install omits the agent label in connector mode" {
    use_podman
    with_tenant

    container_install connector

    [ "$status" -eq 0 ]
    refute_called "--label shellhub.role=agent"
    assert_called "--name=shellhub-connector"
    assert_called "-e SHELLHUB_PRIVATE_KEYS=/host/etc/shellhub/connector/keys"
}

@test "docker_install refuses connector mode without a tenant" {
    container_install connector

    [ "$status" -eq 1 ]
    assert_output_contains "TENANT_ID is required"
    refute_called "docker run"
}

@test "podman_install refuses connector mode without a tenant" {
    use_podman

    container_install connector

    [ "$status" -eq 1 ]
    assert_output_contains "TENANT_ID is required"
    refute_called "podman run"
}

@test "docker_install rejects an unknown mode" {
    container_install bogus

    [ "$status" -eq 1 ]
    assert_output_contains "Invalid mode: bogus"
    refute_called "docker run"
}

@test "podman_install rejects an unknown mode" {
    use_podman

    container_install bogus

    [ "$status" -eq 1 ]
    assert_output_contains "Invalid mode: bogus"
    refute_called "podman run"
}

@test "docker_install maps credentials and preferences onto the agent environment" {
    with_tenant
    export CODE=ABC123 INSTALL_KEY=key-1 KEEPALIVE_INTERVAL=45
    export PREFERRED_HOSTNAME=box PREFERRED_IDENTITY=eth0

    container_install

    [ "$status" -eq 0 ]
    assert_called "-e SHELLHUB_PAIRING_CODE=ABC123"
    assert_called "-e SHELLHUB_INSTALL_KEY=key-1"
    assert_called "-e SHELLHUB_KEEPALIVE_INTERVAL=45"
    assert_called "-e SHELLHUB_PREFERRED_HOSTNAME=box"
    assert_called "-e SHELLHUB_PREFERRED_IDENTITY=eth0"
    assert_called "-e SHELLHUB_TENANT_ID=$TENANT_ID"
    assert_called "-e SHELLHUB_SERVER_ADDRESS=$SERVER_ADDRESS"
}

@test "podman_install maps credentials and preferences onto the agent environment" {
    use_podman
    with_tenant
    export CODE=ABC123 INSTALL_KEY=key-1 KEEPALIVE_INTERVAL=45
    export PREFERRED_HOSTNAME=box PREFERRED_IDENTITY=eth0

    container_install

    [ "$status" -eq 0 ]
    assert_called "-e SHELLHUB_PAIRING_CODE=ABC123"
    assert_called "-e SHELLHUB_INSTALL_KEY=key-1"
    assert_called "-e SHELLHUB_KEEPALIVE_INTERVAL=45"
    assert_called "-e SHELLHUB_PREFERRED_HOSTNAME=box"
    assert_called "-e SHELLHUB_PREFERRED_IDENTITY=eth0"
    assert_called "-e SHELLHUB_TENANT_ID=$TENANT_ID"
    assert_called "-e SHELLHUB_SERVER_ADDRESS=$SERVER_ADDRESS"
}

@test "docker_install omits environment variables for options left unset" {
    container_install

    [ "$status" -eq 0 ]
    refute_called "SHELLHUB_TENANT_ID"
    refute_called "SHELLHUB_PAIRING_CODE"
    refute_called "SHELLHUB_INSTALL_KEY"
    refute_called "SHELLHUB_KEEPALIVE_INTERVAL"
    refute_called "SHELLHUB_PREFERRED_"
    assert_called "-e SHELLHUB_PRIVATE_KEY=$PRIVATE_KEY"
}

@test "podman_install omits environment variables for options left unset" {
    use_podman

    container_install

    [ "$status" -eq 0 ]
    refute_called "SHELLHUB_TENANT_ID"
    refute_called "SHELLHUB_PAIRING_CODE"
    refute_called "SHELLHUB_INSTALL_KEY"
    refute_called "SHELLHUB_KEEPALIVE_INTERVAL"
    refute_called "SHELLHUB_PREFERRED_"
    assert_called "-e SHELLHUB_PRIVATE_KEY=$PRIVATE_KEY"
}

@test "docker_install pulls the agent image by default" {
    container_install

    [ "$status" -eq 0 ]
    assert_called "docker pull -q $AGENT_IMAGE"
    assert_output_contains "Downloading ShellHub container image"
}

@test "podman_install pulls the agent image by default" {
    use_podman

    container_install

    [ "$status" -eq 0 ]
    assert_called "podman pull -q $AGENT_IMAGE"
    assert_output_contains "Downloading ShellHub container image"
}

@test "docker_install skips the pull when the image was overridden" {
    export AGENT_IMAGE_OVERRIDDEN=1

    container_install

    [ "$status" -eq 0 ]
    refute_called "docker pull"
    assert_output_contains "skipping pull"
}

@test "podman_install skips the pull when the image was overridden" {
    use_podman
    export AGENT_IMAGE_OVERRIDDEN=1

    container_install

    [ "$status" -eq 0 ]
    refute_called "podman pull"
    assert_output_contains "skipping pull"
}

@test "docker_install aborts when the image pull fails" {
    stub_bin docker 'echo "docker $*" >> "$CALLS"; [ "$1" = pull ] && exit 1; exit 0'

    container_install

    [ "$status" -eq 1 ]
    assert_output_contains "Failed to download shellhub container image"
    refute_called "docker run"
}

@test "podman_install aborts when the image pull fails" {
    use_podman
    stub_bin podman 'echo "podman $*" >> "$CALLS"; [ "$1" = pull ] && exit 1; exit 0'

    container_install

    [ "$status" -eq 1 ]
    assert_output_contains "Failed to download shellhub container image"
    refute_called "podman run"
}

@test "docker_install honours CONTAINER_NAME" {
    export CONTAINER_NAME=my-agent

    container_install

    [ "$status" -eq 0 ]
    assert_called "--name=my-agent"
    assert_called "docker rm -f my-agent"
}

@test "podman_install honours CONTAINER_NAME" {
    use_podman
    export CONTAINER_NAME=my-agent

    container_install

    [ "$status" -eq 0 ]
    assert_called "--name=my-agent"
}

@test "docker_install removes an existing container so the new one carries the role label" {
    container_install

    [ "$status" -eq 0 ]
    assert_called "docker rm -f shellhub"
}

@test "podman_install checks the boot restart unit before touching podman" {
    RUNTIME=podman

    container_install

    [ "$status" -eq 1 ]
    assert_output_contains "systemctl not found"
    refute_called "podman"
}

@test "install_agent_wrapper writes an executable wrapper bound to podman" {
    call_install install_agent_wrapper podman

    [ "$status" -eq 0 ]
    [ -x "$INSTALL_DIR/shellhub-agent" ]
    assert_output_contains "Installed shellhub-agent wrapper at $INSTALL_DIR/shellhub-agent"
    assert_file_contains "$INSTALL_DIR/shellhub-agent" "podman ps --filter label=shellhub.role=agent"
    assert_file_contains "$INSTALL_DIR/shellhub-agent" "podman exec"
    refute_file_contains "$INSTALL_DIR/shellhub-agent" "docker exec"
}

@test "install_agent_wrapper writes an executable wrapper bound to docker" {
    call_install install_agent_wrapper docker

    [ "$status" -eq 0 ]
    [ -x "$INSTALL_DIR/shellhub-agent" ]
    assert_file_contains "$INSTALL_DIR/shellhub-agent" "docker ps --filter label=shellhub.role=agent"
    assert_file_contains "$INSTALL_DIR/shellhub-agent" "docker exec"
    refute_file_contains "$INSTALL_DIR/shellhub-agent" "podman exec"
}

@test "install_agent_wrapper escalates to write into a root-owned directory" {
    as_non_root

    call_install install_agent_wrapper docker

    [ "$status" -eq 0 ]
    assert_called "sudo tee $INSTALL_DIR/shellhub-agent"
    assert_called "sudo chmod +x $INSTALL_DIR/shellhub-agent"
}

@test "install_agent_wrapper aborts when the install directory does not exist" {
    export INSTALL_DIR="$BATS_TEST_TMPDIR/absent"

    call_install install_agent_wrapper docker

    [ "$status" -eq 1 ]
    assert_output_contains "Failed to install shellhub-agent wrapper"
}

@test "snap_install requires a tenant" {
    stub_bin snap

    call_install snap_install

    [ "$status" -eq 1 ]
    assert_output_contains "TENANT_ID is required"
    refute_called "snap install"
}

@test "snap_install refuses an install key it cannot honour" {
    with_tenant
    export INSTALL_KEY=key-1
    stub_bin snap

    call_install snap_install

    [ "$status" -eq 1 ]
    assert_output_contains "INSTALL_KEY is not supported by the snap install method"
    refute_called "snap install"
}

@test "snap_install aborts when snap is not available" {
    with_tenant

    call_install snap_install

    [ "$status" -eq 1 ]
    assert_output_contains "Snap is not installed or not supported on this system"
}

@test "snap_install configures the service from the environment" {
    with_tenant
    export PREFERRED_HOSTNAME=box KEEPALIVE_INTERVAL=45
    stub_bin snap
    stub_sudo

    call_install snap_install

    [ "$status" -eq 0 ]
    assert_called "snap install --classic --channel=latest/stable shellhub"
    assert_called "snap set shellhub server-address=$SERVER_ADDRESS"
    assert_called "snap set shellhub tenant-id=$TENANT_ID"
    assert_called "snap set shellhub preferred-hostname=box"
    assert_called "snap set shellhub keepalive-interval=45"
    assert_called "snap start shellhub"
}

@test "standalone_install passes the credentials on to the agent service" {
    with_tenant
    export INSTALL_KEY=key-1 KEEPALIVE_INTERVAL=45
    fake_agent_binary

    call_install standalone_install

    [ "$status" -eq 0 ]
    assert_called "agent install --server-address=$SERVER_ADDRESS --tenant-id=$TENANT_ID --install-key=key-1 --keepalive-interval=45"
    [ -x "$INSTALL_DIR/shellhub-agent" ]
}

@test "standalone_install escalates when it is not already root" {
    with_tenant
    as_non_root
    fake_agent_binary

    call_install standalone_install

    [ "$status" -eq 0 ]
    assert_output_contains "This install method requires root privileges"
    assert_called "sudo $INSTALL_DIR/shellhub-agent install"
}

@test "standalone_install removes the binary when the service install fails" {
    with_tenant
    fake_agent_binary 1

    call_install standalone_install

    [ "$status" -eq 1 ]
    assert_output_contains "Failed to install ShellHub agent service"
    [ ! -e "$INSTALL_DIR/shellhub-agent" ]
}

@test "docker_uninstall removes the container and the wrapper" {
    stub_bin docker
    call_install install_agent_wrapper docker

    call_install docker_uninstall

    [ "$status" -eq 0 ]
    assert_called "docker rm -f shellhub"
    [ ! -e "$INSTALL_DIR/shellhub-agent" ]
}

@test "docker_uninstall escalates when it is not already root" {
    stub_bin docker
    call_install install_agent_wrapper docker
    as_non_root

    call_install docker_uninstall

    [ "$status" -eq 0 ]
    assert_called "sudo docker rm -f shellhub"
    assert_called "sudo rm -f $INSTALL_DIR/shellhub-agent"
}

@test "podman_uninstall removes the container and the wrapper" {
    stub_bin podman
    call_install install_agent_wrapper podman

    call_install podman_uninstall

    [ "$status" -eq 0 ]
    assert_called "podman rm -f shellhub"
    [ ! -e "$INSTALL_DIR/shellhub-agent" ]
}

@test "podman_uninstall escalates when it is not already root" {
    stub_bin podman
    call_install install_agent_wrapper podman
    as_non_root

    call_install podman_uninstall

    [ "$status" -eq 0 ]
    assert_called "sudo podman rm -f shellhub"
    assert_called "sudo rm -f $INSTALL_DIR/shellhub-agent"
}

@test "docker_uninstall reports a container that was already gone" {
    stub_bin docker 'echo "docker $*" >> "$CALLS"; exit 1'

    call_install docker_uninstall

    [ "$status" -eq 0 ]
    assert_output_contains "not found (may already be removed)"
}

@test "standalone_uninstall reports a missing binary" {
    call_install standalone_uninstall

    [ "$status" -eq 1 ]
    assert_output_contains "ShellHub agent binary not found"
}

@test "standalone_uninstall stops the service and removes the binary" {
    with_tenant
    fake_agent_binary
    call_install standalone_install

    call_install standalone_uninstall

    [ "$status" -eq 0 ]
    assert_called "agent uninstall"
    [ ! -e "$INSTALL_DIR/shellhub-agent" ]
}

@test "wsl_install aborts when systemd is not enabled" {
    stub_bin systemctl 'exit 1'

    call_install wsl_install

    [ "$status" -eq 1 ]
    assert_output_contains "requires systemd to be enabled"
}

@test "wsl_install aborts when WSL networking is not mirrored" {
    stub_bin systemctl 'exit 0'
    stub_bin wslinfo 'echo nat'

    call_install wsl_install

    [ "$status" -eq 1 ]
    assert_output_contains "networking mode must be set to mirrored"
}

@test "wsl_install falls through to the standalone install once both preconditions hold" {
    with_tenant
    stub_bin systemctl 'exit 0'
    stub_bin wslinfo 'echo mirrored'
    fake_agent_binary

    call_install wsl_install

    [ "$status" -eq 0 ]
    assert_called "agent install --server-address=$SERVER_ADDRESS"
}

@test "the installer aborts on FreeBSD" {
    stub_bin uname 'echo FreeBSD'

    run_install

    [ "$status" -eq 1 ]
    assert_output_contains "Automatic installation is not supported on FreeBSD"
}

@test "detection prefers docker when it is accessible" {
    stub_bin docker
    stub_bin podman

    run_install

    [ "$status" -eq 0 ]
    assert_output_contains "Installing ShellHub using docker method"
    assert_called "docker run -d"
    refute_called "podman run"
}

@test "detection falls back to podman when docker is absent" {
    stub_bin podman
    stub_bin systemctl 'exit 0'

    run_install

    [ "$status" -eq 0 ]
    assert_output_contains "Installing ShellHub using podman method"
    assert_called "podman run -d"
}

@test "detection retries docker with sudo when it is only reachable as root" {
    unset AGENT_IMAGE
    stub_bin docker 'echo "docker $*" >> "$CALLS"; [ "$1" = info ] && [ -z "$RAN_UNDER_SUDO" ] && exit 1; exit 0'
    stub_sudo

    run_install

    [ "$status" -eq 0 ]
    assert_called "sudo docker info"
    assert_called "sudo docker pull -q docker.io/shellhubio/agent:$AGENT_VERSION"
    assert_called "sudo docker run -d"
}

@test "detection retries podman with sudo when it is only reachable as root" {
    unset AGENT_IMAGE
    stub_bin podman 'echo "podman $*" >> "$CALLS"; [ "$1" = info ] && [ -z "$RAN_UNDER_SUDO" ] && exit 1; exit 0'
    stub_bin systemctl 'exit 0'
    stub_sudo

    run_install

    [ "$status" -eq 0 ]
    assert_called "sudo podman info"
    assert_called "sudo podman pull -q docker.io/shellhubio/agent:$AGENT_VERSION"
    assert_called "sudo podman run -d"
}

@test "detection falls back to snap when no container runtime is available" {
    with_tenant
    stub_bin snap
    stub_sudo

    run_install

    [ "$status" -eq 0 ]
    assert_output_contains "Installing ShellHub using snap method"
    assert_called "snap install --classic --channel=latest/stable shellhub"
}

@test "detection falls back to the standalone install when nothing else is available" {
    fake_agent_binary

    run_install

    [ "$status" -eq 0 ]
    assert_output_contains "Installing ShellHub using standalone method"
    assert_called "agent install --server-address=$SERVER_ADDRESS"
}

@test "an explicit INSTALL_METHOD skips detection entirely" {
    export INSTALL_METHOD=standalone
    stub_bin docker
    fake_agent_binary

    run_install

    [ "$status" -eq 0 ]
    assert_output_contains "Installing ShellHub using standalone method"
    refute_called "docker info"
}

@test "an unsupported INSTALL_METHOD is rejected" {
    export INSTALL_METHOD=bogus

    run_install

    [ "$status" -eq 1 ]
    assert_output_contains "Install method not supported"
}

@test "detection reports WSL 2 on Ubuntu" {
    enter_wsl
    stub_bin systemctl 'exit 0'
    stub_bin wslinfo 'echo mirrored'
    fake_agent_binary

    run_install

    [ "$status" -eq 0 ]
    assert_output_contains "Detected WSL environment"
    assert_output_contains "Installing ShellHub using WSL method"
    assert_called "agent install --server-address=$SERVER_ADDRESS"
}

@test "detection aborts when WSL is older than version 2" {
    enter_wsl 1.2.5.0

    run_install

    [ "$status" -eq 1 ]
    assert_output_contains "WSL version 2 is required"
}

@test "detection aborts on WSL with a distro other than Ubuntu" {
    enter_wsl
    echo 'NAME="Debian GNU/Linux"' > "$OS_RELEASE"

    run_install

    [ "$status" -eq 1 ]
    assert_output_contains "Only Ubuntu is supported in WSL"
}

@test "WSL detection overrides an explicit INSTALL_METHOD, so a chosen method does not survive WSL" {
    export INSTALL_METHOD=standalone
    enter_wsl
    stub_bin systemctl 'exit 0'
    stub_bin wslinfo 'echo mirrored'
    fake_agent_binary

    run_install

    [ "$status" -eq 0 ]
    assert_output_contains "Installing ShellHub using WSL method"
}

@test "the installer reports the enrollment credential before installing anything" {
    with_tenant
    stub_bin docker

    run_install

    [ "$status" -eq 0 ]
    assert_output_contains "Enrollment: tenant $TENANT_ID (device lands pending)"
}

@test "uninstall dispatches to the detected method" {
    stub_bin docker

    run_install uninstall

    [ "$status" -eq 0 ]
    assert_output_contains "Uninstalling ShellHub using docker method"
    assert_called "docker rm -f shellhub"
}

@test "uninstall is refused for install methods that do not support it" {
    export INSTALL_METHOD=snap

    run_install uninstall

    [ "$status" -eq 1 ]
    assert_output_contains "Uninstall is not yet supported for 'snap' install method"
}
