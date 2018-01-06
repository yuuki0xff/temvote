#!/bin/bash
# start-setup.sh /path/to/root cmd [args]
# USBメモリ内に含まれるファイルの署名を検証後、引数で指定されたコマンドの実行を開始する。
set -eu

TRUSTED_KEYS_DIR=/srv/deploy/trusted.gpg.d
TRUSTED_KEYS_DIR=$PWD/gpg
USB_MEMORY_ROOT=$1
shift

function generate_sha256sum() {
    cd "${USB_MEMORY_ROOT}"
    find -type f \
         ! '(' -name sha256sum -o -name sha256sum.sig -o -name '*.list' ')' \
         -exec sha256sum '{}' '+' |sort
}

# find gpg command
if which gpg; then GPG=gpg
elif which gpgv; then GPG=gpgv
elif which gpgv1; then GPG=gpgv1
elif which gpgv2; then GPG=gpgv2
else
    echo "ERROR: Not found gpg command" >&2
    eixt 1
fi

echo
echo === Actual file hashes ===
generate_sha256sum
echo
echo === Expected file hashes ===
cat "${USB_MEMORY_ROOT}/sha256sum"
echo

if ! diff "${USB_MEMORY_ROOT}/sha256sum" <(generate_sha256sum); then
    echo "ERROR: SOME FILEs ARE TAMPERED !" >&2
    exit 1
fi

trusted=0
for keyfile in ${TRUSTED_KEYS_DIR}/*; do
    if $GPG -v --homedir /tmp \
        --no-default-keyring \
        --keyring "$keyfile" \
        --verify "${USB_MEMORY_ROOT}/sha256sum.sig"; then
        trusted=1
        break
    fi
done

if [ "$trusted" -eq 0 ]; then
    echo "ERROR: The USB memory signature IS NOT signed by trusted keys." >&2
    exit 1
fi

"$@"
