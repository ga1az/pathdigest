#!/bin/sh
set -e

usage() {
  this=$1
  cat <<EOF
$this: Install pathdigest

Usage: $this [-b bindir] [-d] [tag]
  -b bindir    Installation directory. Defaults to $(go env GOPATH)/bin if available, otherwise ./bin
  -d           Activate debug logs.
   [tag]       A release tag of https://github.com/ga1az/pathdigest/releases
               If omitted, the latest version will be used.

EOF
  exit 2
}

parse_args() {
  if go env GOPATH >/dev/null 2>&1; then
    DEFAULT_BINDIR=$(go env GOPATH)/bin
  else
    DEFAULT_BINDIR=./bin
  fi
  BINDIR=${BINDIR:-${DEFAULT_BINDIR}}

  while getopts "b:dh?x" arg; do
    case "$arg" in
      b) BINDIR="$OPTARG" ;;
      d) log_set_priority 10 ;;
      h | \?) usage "$0" ;;
      x) set -x ;;
    esac
  done
  shift $((OPTIND - 1))
  TAG=$1
}

execute() {
  tmpdir=$(mktemp -d)
  log_debug "Downloading files to ${tmpdir}"
  http_download "${tmpdir}/${TARBALL_FILENAME}" "${TARBALL_URL}"
  http_download "${tmpdir}/${CHECKSUM_FILENAME}" "${CHECKSUM_URL}"
  hash_sha256_verify "${tmpdir}/${TARBALL_FILENAME}" "${tmpdir}/${CHECKSUM_FILENAME}"
  
  log_debug "Extracting ${TARBALL_FILENAME} in ${tmpdir}"
  (cd "${tmpdir}" && untar "${TARBALL_FILENAME}")

  SOURCE_BIN_PATH="${tmpdir}/${BINARY_NAME}"
  if [ "$OS" = "windows" ]; then
    SOURCE_BIN_PATH="${tmpdir}/${BINARY_NAME}.exe"
  fi

  if [ ! -f "$SOURCE_BIN_PATH" ]; then
    log_crit "Binary ${BINARY_NAME} not found in the downloaded file."
    exit 1
  fi

  test ! -d "${BINDIR}" && install -d "${BINDIR}"
  
  TARGET_BIN_PATH="${BINDIR}/${BINARY_NAME}"
  if [ "$OS" = "windows" ]; then
    TARGET_BIN_PATH="${BINDIR}/${BINARY_NAME}.exe"
  fi

  install "${SOURCE_BIN_PATH}" "${TARGET_BIN_PATH}"
  log_info "Installed ${TARGET_BIN_PATH}"
  
  rm -rf "${tmpdir}"
}

get_binaries_info() {
  OS=$(uname_os)
  ARCH=$(uname_arch)
  PLATFORM="${OS}/${ARCH}"

  BINARY_NAME="pathdigest"

  ARCHIVE_FORMAT="tar.gz"
  if [ "$OS" = "windows" ]; then
    ARCHIVE_FORMAT="zip"
  fi

  case "$PLATFORM" in
    darwin/amd64) ;;
    darwin/arm64) ;;
    linux/386) ;;
    linux/amd64) ;;
    linux/arm64) ;;
    windows/386) ;;
    windows/amd64) ;;
    windows/arm64) ;;
    *)
      log_crit "Platform ${PLATFORM} is not supported by this installation script or there are no precompiled binaries for it."
      log_crit "Ensure this script is up to date and/or request binaries at https://github.com/${OWNER}/${REPO}/issues/new"
      exit 1
      ;;
  esac
}

PROJECT_NAME="pathdigest"
OWNER="ga1az"
REPO="pathdigest"

GITHUB_REPO_SLUG="${OWNER}/${REPO}"

log_prefix() {
  echo "$PROJECT_NAME installer"
}

uname_os_check
uname_arch_check

parse_args "$@"

get_binaries_info

tag_to_version

log_info "Version found: ${VERSION} for ${TAG} in ${PLATFORM}"

ARTEFACT_BASE_NAME="${PROJECT_NAME}_${VERSION}_${OS}_${ARCH}"
ARTEFACT_BASE_NAME="${PROJECT_NAME}_${VERSION}_${OS}_${ARCH}"
TARBALL_FILENAME="${ARTEFACT_BASE_NAME}.${ARCHIVE_FORMAT}"
CHECKSUM_FILENAME="${PROJECT_NAME}_${VERSION}_checksums.txt"

GITHUB_DOWNLOAD_URL="https://github.com/${GITHUB_REPO_SLUG}/releases/download/${TAG}"
TARBALL_URL="${GITHUB_DOWNLOAD_URL}/${TARBALL_FILENAME}"
CHECKSUM_URL="${GITHUB_DOWNLOAD_URL}/${CHECKSUM_FILENAME}"

execute
