#!/bin/bash

ROOT=$(dirname "$PWD")
BUILDOUT=$ROOT/build/bin/
mkdir -p $BUILDOUT
GAMMAOUT=$ROOT/build/gamma_build

# BUILD OPTS
COMPILE_THREAD_NUM=-j2
BUILD_GAMMA=ON
BUILD_GAMMA_TEST=OFF
BUILD_GAMMA_TYPE=Release

# version value
BUILD_VERSION="latest"

while getopts ":n:g:tdh" opt; do
  case $opt in
  n)
    COMPILE_THREAD_NUM="-j"$OPTARG
    echo "COMPILE_THREAD_NUM="$COMPILE_THREAD_NUM
    ;;
  t)
    BUILD_GAMMA_TEST=ON
    echo "BUILD_GAMMA_TEST=ON"
    ;;
  d)
    BUILD_GAMMA_TYPE=Debug
    echo "BUILD_GAMMA_TYPE="$BUILD_GAMMA_TYPE
    ;;
  g)
    BUILD_GAMMA=$OPTARG
    echo "BUILD_GAMMA="$BUILD_GAMMA
    ;;
  h)
    echo "[build options]"
    echo -e "\t-h\t\thelp"
    echo -e "\t-n\t\tcompile thread num"
    echo -e "\t-g\t\tbuild gamma or not: [ON|OFF]"
    echo -e "\t-t\t\tbuild gamma test"
    echo -e "\t-d\t\tbuild gamma type=Debug"
    exit 0
    ;;
  ?)
    echo "unsupport param, -h for help"
    exit 1
    ;;
  esac
done

ROCKSDB_URL=https://github.com/facebook/rocksdb/archive/refs/tags/v6.6.4.tar.gz

function get_version() {
  VEARCH_VERSION_MAJOR=$(cat ${ROOT}/VERSION | grep VEARCH_VERSION_MAJOR | awk -F' ' '{print $2}')
  VEARCH_VERSION_MINOR=$(cat ${ROOT}/VERSION | grep VEARCH_VERSION_MINOR | awk -F' ' '{print $2}')
  VEARCH_VERSION_PATCH=$(cat ${ROOT}/VERSION | grep VEARCH_VERSION_PATCH | awk -F' ' '{print $2}')

  BUILD_VERSION="v${VEARCH_VERSION_MAJOR}.${VEARCH_VERSION_MINOR}.${VEARCH_VERSION_PATCH}"
  echo "BUILD_VERSION="${BUILD_VERSION}
}

function build_thirdparty() {
  OS_NAME=$(uname)
  if [ ${OS_NAME} == "Darwin" ]; then
    export ROCKSDB_HOME=/usr/local/include/rocksdb
    brew install rocksdb
  else
    if [ ! -n "${ROCKSDB_HOME}" ]; then
      export ROCKSDB_HOME=/usr/local/include/rocksdb
      if [ ! -d "${ROCKSDB_HOME}" ]; then
        rm -rf rocksdb*
        wget ${ROCKSDB_URL} -O rocksdb.tar.gz
        tar -xzf rocksdb.tar.gz
        pushd rocksdb-6.6.4
        CFLAGS="-O3 -fPIC" make shared_lib $COMPILE_THREAD_NUM
        make install
        popd
      fi
    fi
  fi
}

function build_engine() {
  echo "build gamma"
  rm -rf ${GAMMAOUT} && mkdir -p $GAMMAOUT
  pushd $GAMMAOUT
  cmake -DPERFORMANCE_TESTING=ON -DCMAKE_BUILD_TYPE=$BUILD_GAMMA_TYPE -DBUILD_TEST=$BUILD_GAMMA_TEST $ROOT/engine/
  make $COMPILE_THREAD_NUM
  popd
}

function build_vearch() {
  flags="-X 'main.BuildVersion=$BUILD_VERSION' -X 'main.CommitID=$(git rev-parse HEAD)' -X 'main.BuildTime=$(date +"%Y-%m-%d %H:%M.%S")'"
  echo "version info: $flags"
  export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:$GAMMAOUT
  export LIBRARY_PATH=$LIBRARY_PATH:$GAMMAOUT

  echo "build vearch"
  go build -a -tags="vector" -ldflags "$flags" -o $BUILDOUT/vearch $ROOT/startup.go
  echo "build deploy tool"
  go build -a -ldflags "$flags" -o $BUILDOUT/batch_deployment $ROOT/tools/deployment/batch_deployment.go
}

get_version
if [ $BUILD_GAMMA == "ON" ]; then
  build_thirdparty
  build_engine
fi
build_vearch
