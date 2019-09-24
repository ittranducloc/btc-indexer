#!/usr/bin/env bash

SCRIPT_NAME=$(basename "$0")
BUILD_DIR=$(dirname "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)")

function usage() {
  echo ""
  echo "Usage: $SCRIPT_NAME [OPTIONS]"
  echo ""
  echo "Build project services"
  echo ""
  echo "Options:"
  echo "  -h | --help"
  echo "  -c | --cache"
  echo "  -s | --single"
  echo "  -v | --verbose"
}

function build_single() {
  APP_NAME=$(basename "$1")
  echo -e ">> \e[32;5mbuilding $APP_NAME ...\e[0m"
  cd "$1" || exit
  time go build -ldflags "-s -w" || return $?
  ls -lh "$APP_NAME"
}

function build() {
  ok_nbr=0
  failed_nbr=0
  for APP_PATH in "$BUILD_DIR"/cmd/* ; do
    build_single "$APP_PATH"
    ERR=$?
    if [[ $ERR != 0 ]]; then
      ((failed_nbr++))
    else
      ((ok_nbr++))
    fi
  done

  if [[ $failed_nbr != 0 ]]; then
    echo -e "failure happened, \e[32;5msucceed $ok_nbr\e[0m \e[31;5mfailed $failed_nbr\e[0m"
    return 1
  else
    echo -e "build completed, \e[32;5msucceed $ok_nbr\e[0m"
    return 0
  fi
}

function main() {
  while true; do
    case $1 in
    -h | --help)
      HELP=true
      shift
      ;;
    -c | --cache)
      BUILD_WITH_CACHE=true
      shift
      ;;
    -s | --single)
      BUILD_SINGLE_APP=true
      APP=$2
      shift
      ;;
    -v | --verbose)
      VERBOSE=true
      shift
      ;;
    --)
      shift
      break
      ;;
    *) break ;;
    esac
  done

  if [[ $HELP == true ]]; then
    usage
    exit 0
  fi

  if [[ $BUILD_WITH_CACHE == true ]]; then
    # Cache in the folder '.idea' to prevent Git to index
    export GOCACHE=${BUILD_DIR}/.idea/cache/go-build
    mkdir -p "$GOCACHE"
  fi

  if [[ $VERBOSE == true ]]; then
    echo "Build Dir:        $BUILD_DIR"
    echo "Build With Cache: $BUILD_WITH_CACHE"
    echo "Go Version:       $(go version)"
    echo "Build Environment:"
    go env
    echo ""
  fi

  if [[ $BUILD_SINGLE_APP == true ]]; then
    APP_PATH=$(ls -d "$BUILD_DIR"/cmd/* | grep "$APP")
    if [[ $APP_PATH == "" ]]; then
      echo "Not found the input app: $APP"
      exit 1
    fi
    build_single "$APP_PATH"
    exit $?
  fi

  build
  exit $?
}

main "$@"
