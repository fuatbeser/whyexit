#!/bin/sh
# Install the latest whyexit release:
#
#   curl -fsSL https://raw.githubusercontent.com/fuatbeser/whyexit/main/install.sh | sh
#
# WHYEXIT_VERSION picks a release (e.g. 0.1.0) and WHYEXIT_INSTALL_DIR
# the target directory.
set -eu

repo=fuatbeser/whyexit
base=${WHYEXIT_DOWNLOAD_URL:-https://github.com/$repo/releases}

fail() {
	echo "install.sh: $*" >&2
	exit 1
}

os=$(uname -s | tr '[:upper:]' '[:lower:]')
case $os in
linux | darwin | freebsd) ;;
*) fail "unsupported OS: $os" ;;
esac

case $(uname -m) in
x86_64 | amd64) arch=amd64 ;;
aarch64 | arm64) arch=arm64 ;;
*) fail "unsupported CPU: $(uname -m)" ;;
esac

if [ -n "${WHYEXIT_VERSION:-}" ]; then
	url=$base/download/v${WHYEXIT_VERSION#v}
else
	url=$base/latest/download
fi

dir=${WHYEXIT_INSTALL_DIR:-}
if [ -z "$dir" ]; then
	if [ -w /usr/local/bin ]; then
		dir=/usr/local/bin
	else
		dir=$HOME/.local/bin
	fi
fi

if command -v sha256sum >/dev/null; then
	sha256="sha256sum"
elif command -v shasum >/dev/null; then
	sha256="shasum -a 256"
else
	fail "need sha256sum or shasum to verify the download"
fi

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

archive=whyexit_${os}_${arch}.tar.gz
echo "Downloading $url/$archive"
curl -fsSL -o "$tmp/$archive" "$url/$archive" || fail "download failed"
curl -fsSL -o "$tmp/checksums.txt" "$url/checksums.txt" || fail "checksum download failed"

(cd "$tmp" && grep " $archive\$" checksums.txt | $sha256 -c - >/dev/null) ||
	fail "checksum mismatch for $archive"

tar -xzf "$tmp/$archive" -C "$tmp" whyexit
mkdir -p "$dir"
mv "$tmp/whyexit" "$dir/whyexit"
chmod 755 "$dir/whyexit"
echo "Installed $("$dir/whyexit" --version) to $dir"

case :$PATH: in
*:"$dir":*) ;;
*) echo "Add $dir to your PATH to use it." ;;
esac
