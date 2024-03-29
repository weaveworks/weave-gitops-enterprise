#!/usr/bin/env bash
# shellcheck shell=bash

set -eo pipefail

unset CD_PATH
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "${SCRIPT_DIR}" || exit 1

# shellcheck source=./functions.sh
. "${SCRIPT_DIR}/functions.sh"

DEP_FILE=${1}
BIN_DIR=${2:-${SCRIPT_DIR}/bin}
CACHE_DIR="${SCRIPT_DIR}/cache"
RELEASE_GOOS=${RELEASE_GOOS:-$(go env GOOS)}
SKIP_FETCH_TOOLS=${SKIP_FETCH_TOOLS:-""}

# Derive from GOOS
RELEASE_OS=$(title_case "$RELEASE_GOOS")

if [ -n "$SKIP_FETCH_TOOLS" ]; then
    echo "skipping fetch tools..."
    exit 0
fi

# just in case we're bootstrapping, make sure the cache exists
mkdir -p "${CACHE_DIR}"

# create bin directory
rm -rf "${BIN_DIR}"
mkdir "${BIN_DIR}"

# add binaries

# Check if a url points to a valid location
check_url() {
    test $# == 1 && test "$1" || return 1
    curl --output /dev/null --silent --head --fail "$1"
}

instantiate_url() {
    test $# == 1 || exit
    local url="${1}"
    url=${url//\$\{arch\}/$(arch)}
    url=${url//\$\{goarch\}/$(goarch)}
    url=${url//\$\{goos\}/$RELEASE_GOOS}
    url=${url//\$\{os\}/$RELEASE_OS}
    url=${url//\$\{version\}/$(run_stoml version)}
    # hack for tilt because of https://github.com/tilt-dev/tilt/issues/5434
    if [[ "${url}" = *"tilt"* ]] && [[ "${url}" = *"darwin"* ]]; then
      url="${url/darwin/mac}"
    fi
    if [[ "${url}" = *"github.com/cli/cli"* ]] && [[ "${url}" = *"darwin"* ]]; then
      url="${url//darwin/macOS}"
    fi
    echo "${url}"
}

# select either binary or tar download by checking for existence
# and allow overriding with a local tool by setting the environment variable "$LOCAL_<tool>" (e.g. $LOCAL_wk)
# (useful for testing and particularly useful on darwin since we don't publish a "wk" version for darwin)
download_dependency() {
    local tool="${1}"
    local bin_dir="${2}"
    local dependencies_toml="${DEP_FILE}"

    # short circuit if a LOCAL_${tool} value is in the env
    local localToolVar="\$LOCAL_"${tool//-/_}
    local localTool
    localTool=$(eval "echo ${localToolVar}")
    if [ -n "${localTool}" ]; then
        echo "Using ${tool} given in ${localToolVar} at ${localTool}"
        cp "${localTool}" "${bin_dir}"
        return 0
    fi

    run_stoml() {
        local property="${1}"
        "${bin_dir}"/stoml "${dependencies_toml}" "${tool}"."${property}"
    }

    local version="$(run_stoml version)"

    # short circuit if we already have the tool at the version in the cache
    local cached=${tool}_${version}
    if [ -e "${CACHE_DIR}/${cached}" ]; then
        echo "Using ${tool} ${version} in cache at ${CACHE_DIR}/${cached}"
        cp -r "${CACHE_DIR}/${cached}" "${custom_bindir:-$bin_dir}/${tool}"
        return 0
    fi

    local tarpath
    tarpath=$(instantiate_url "$(run_stoml tarpath)")

    local special_tarpath
    special_tarpath=$(instantiate_url "$(run_stoml special_tarpath)")

    local special_tarpath_url
    # shellcheck disable=SC2206
    special_tarpath_url=(${special_tarpath//;/ }) # split out special paths which contain <url>;<path in tarball>

    local binarypath
    binarypath=$(instantiate_url "$(run_stoml binarypath)")

    local checksum_path
    local url_and_path
    local txtpath
    txtpath=$(instantiate_url "$(run_stoml txtpath)")

    local custom_bindir
    custom_bindir=$(run_stoml bindir)
    mkdir -p "${custom_bindir:-$bin_dir}"

    if check_url "${txtpath}"; then
        url_and_path="${txtpath}"
        checksum_path="${custom_bindir}/${tool}_checksum.txt"

        do_curl "${checksum_path}" "${url_and_path}"
    fi

    local fetch
    # shellcheck disable=SC2128
    if check_url "${binarypath}"; then
        url_and_path="${binarypath}"
        fetch=do_curl_binary
    elif check_url "${special_tarpath_url}"; then
        url_and_path="${special_tarpath}"
        fetch=do_curl_tarball_with_path
    elif check_url "${tarpath}"; then
        url_and_path="${tarpath}"
        fetch=do_curl_tarball
    else
        echo "No valid path for tool:" "${tool}"
        exit 1
    fi

    "${fetch}" "${cached}" "${url_and_path}" "${CACHE_DIR}" "${checksum_path}"
    cp -r "${CACHE_DIR}/${cached}" "${custom_bindir:-$bin_dir}/${tool}"
}

get_tool() {
    local tool="${1}"
    local url="${3}"
}

# Don't use $RELEASE_GOOS here, should be whatever is running the script.
stoml_version="0.4.0"
stoml_url="https://github.com/freshautomations/stoml/releases/download/v${stoml_version}/stoml_$(goos)_amd64"
stoml_cached=stoml_${stoml_version}
if [ ! -e "${CACHE_DIR}/${stoml_cached}" ]; then
    do_curl_binary "${stoml_cached}" "${stoml_url}" "${CACHE_DIR}"
else
    echo "Using stoml ${stoml_version} in cache at ${CACHE_DIR}/${stoml_cached}"
fi
cp -r "${CACHE_DIR}/${stoml_cached}" "${BIN_DIR}/stoml"

# Downloading tools
tools=$("${BIN_DIR}"/stoml "${DEP_FILE}" .)
for tool in $tools; do
    download_dependency "${tool}" "${BIN_DIR}"
done
