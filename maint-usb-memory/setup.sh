#!/usr/bin/env bash
set -eu

exec 20>>/var/tmp/temvote-setup.$(date +%04Y%02m%02d%02H%02M%02S).log
export BASH_XTRACEFD=20
set -x

SELF=$(readlink -f "$0")
SELF_DIR=$(dirname "$SELF")
. "${SELF_DIR}/setup.conf"

SSH_PRV_KEY=${SELF_DIR}/ssh/id_rsa.pub
SSH_JUMP_HOST=${SSH_JUMP_HOST:-}
REMOTE_SSH_SOCKET=${REMOTE_SSH_SOCKET:-}

if [ -z "$SSH_JUMP_HOST" ]; then
    echo 'ERROR: Should set $SSH_JUMP_HOST=user@host'
    exit 1
fi >&2
if [ -z "$REMOTE_SSH_SOCKET" ]; then
    echo 'ERROR: Should set $REMOTE_SSH_SOCKET'
    exit 1
fi >&2

export TEMVOTE_NOREBOOT="${TEMVOTE_NOREBOOT:-}"
export DEBIAN_FRONTEND=noninteractive
export AUTOSSH_POLL=20

SETUP_SH_HELP="
Usage: sudo $SELF bme280d-setup-base
       sudo $SELF bme280d-setup
       sudo $SELF bme280d-debug

For Manually Operation:
  sudo $SELF setup_new_node <hostname> <user> <password>
  sudo $SELF install_wifi_config
  sudo $SELF wait_internet_access
  sudo $SELF upgrade_all_packages
  sudo $SELF install_ntpdated
  sudo $SELF install_bme280d_service
  sudo $SELF enable_i2c
  sudo $SELF control_debug_services <start|stop|enable|disable>
  sudo $SELF start_ssh_tunnel
  sudo $SELF system_reset <reboot|poweroff>
"

NEW_NODE_MSG="
================================

hostname:       %s
login user:     %s
login password: %s

================================
"

function _random_hex_str() {
    local length=$1
    dd if=/dev/urandom bs="$length" count=1 |od -x -A none |tr -d ' ' |head --bytes="$length"
}

# 英数字と記号混じりの文字列を出力する
# パスワードなどに最適
function _random_str() {
    local length=$1
    dd if=/dev/urandom bs="$length" count=1 |base64 |tr -d "\r\n" |head --bytes="$length"
}

function _atomic_append() {
    local file=$1
    local content=$2

    local backup_file="${file}.old"
    local tmp_file="${file}.tmp"

    # create a backup file
    rm -f "$backup_file"
    ln "$file" "$backup_file"

    # create a temporary file
    cp -a "$file" "$tmp_file"
    echo "$content" >>"$tmp_file"
    sync --file-system "$tmp_file"

    # atomically overwrite to $file
    mv "$tmp_file" "$file"
    sync --file-system "$file"
}

function _change_hostname() {
    local new_hostname=$1
    echo -n "${new_hostname}" >/etc/hostname
    hostname --file /etc/hostname
}

function _change_user_password() {
    local user_name=$1
    local user_password=$2
    echo "${user_name}:${user_password}" |chpasswd
}

function setup_new_node() {
    local new_hostname=temvote-$(_random_hex_str)
    local user_name=pi
    local user_password=$(_random_str)

    _change_hostname "$new_hostname"
    _change_user_password "$user_name" "$user_password"

    # save hostname and password into USB memory
    lcaol pass_file="${SELF_DIR}/host_passwd.list"
    local content="${new_hostname}:${user_name}:${user_password}"
    _atomic_append "$pass_file" "$content"

    printf "$NEW_NODE_MSG" \
        "$new_hostname" \
        "$user_name" \
        "$user_password"
}

function install_wifi_config() {
    install -Cd -o root -g root -m 700 /etc/wpa_supplicant
    install -C  -o root -g root -m 600 "${SELF_DIR}/config/wpa_supplicant.conf" /etc/wpa_supplicant/wpa_supplicant-wlan0.conf
    systemctl restart wpa_supplicant.service
    systemctl restart networking.service
    systemctl enable wpa_supplicant@wlan0.service
    systemctl restart wpa_supplicant@wlan0.service
}

function wait_internet_access() {
    until curl --connect-timeout 5 http://example.com &>/dev/null; do
        sleep 5
    done
}

function upgrade_all_packages() {
    sudo apt update
    # may be disconnected from the Internet when firmware updated.
    # so, all package downloads before upgraded.
    sudo apt upgrade -y --download-only
    sudo apt install -y --download-only ntpdate autossh python3 python3-pip

    sudo apt upgrade -y
    sudo apt install -y ntpdate autossh python3 python3-pip
    sudo python3 -m pip install -r "$SELF_DIR/service/requirements.txt"
}

function install_ntpdated() {
    install -C  -o root -g root   -m 755 "${SELF_DIR}/service/ntpdated.service" /etc/systemd/system/ntpdated.service
    install -C  -o root -g root   -m 755 "${SELF_DIR}/service/ntpdated.timer" /etc/systemd/system/ntpdated.timer
    systemctl daemon-reload
    systemctl enable ntpdated.service ntpdated.timer
}

function install_bme280d_service() {
    install -Cd -o root -g root   -m 755 /srv
    install -Cd -o root -g root   -m 755 /srv/bme280d/
    install -Cd -o root -g root   -m 755 /srv/bme280d/bin/
    install -Cd -o root -g daemon -m 750 /srv/bme280d/conf/
    install -C  -o root -g daemon -m 750 "${SELF_DIR}/service/bme280d" /srv/bme280d/bin/bme280d
    install -C  -o root -g root   -m 600 "${SELF_DIR}/service/bme280d.service" /etc/systemd/system/bme280d.service
    install -C  -o root -g root   -m 600 "${SELF_DIR}/service/bme280d.conf" /srv/bme280d/conf/bme280d.conf
    systemctl daemon-reload
    systemctl enable bme280d.service
}

function enable_i2c() {
    # NOTE: 0 is enable
    #       1 is disable
    raspi-config nonint do_i2c 0
}

function control_debug_services() {
    # action is "start", "stop", "enable" or "disable"
    local action=$1
    systemctl "$action" \
        ssh.service \
        getty@tty1.service \
        getty@tty2.service \
        getty@tty3.service \
        getty@tty4.service \
        getty@tty5.service
}

function start_ssh_tunnel() {
    autossh -R "${REMOTE_SSH_SOCKET}:localhost:22" "$SSH_JUMP_HOST"
}

function system_reset() {
    local action=$1
    local message=$2

    if [ -z "${TEMVOTE_NOREBOOT:-}" ]; then
        cp -a "${SELF_DIR}/system-reset.sh" "/tmp"
        exec bash /tmp/system-reset.sh "$action" "$SELF_DIR" "$message"
    fi
    exit
}

if [ $# = 0 ]; then
    echo -n "$SETUP_SH_HELP"
    exit 1
fi
case "$1" in
    --help|help)
        echo -n "$SETUP_SH_HELP"
        ;;
    bme280d-setup-base)
        # setup without bme280d service
        install_wifi_config
        wait_internet_access
        upgrade_all_packages
        enable_i2c
        system_reset poweroff "Finished basic setup."
        ;;
    bme280d-setup)
        # install configurations, bme280d service, and upgrade installed packages.
        install_wifi_config
        wait_internet_access
        install_ntpdated
        install_bme280d_service
        upgrade_all_packages
        enable_i2c
        control_debug_services disable
        system_reset reboot "Finished all setup."
        ;;
    bme280d-debug)
        control_debug_services start
        start_ssh_tunnel
        system_reset reboot "When finish debugging, please disconnect the USB memory."
        ;;
    *)
        # execute a defined function manually
        "$@"
        ;;
esac

