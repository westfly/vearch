#!/usr/bin/env bash
#add env
export GOROOT=/env/app/go
export PATH=$PATH:$GOROOT/bin
export ZFP_HOME=/env/app/zfp_install
export ROCKSDB_HOME=/env/app/rocksdb_install
export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:$ROCKSDB_HOME/lib:$ZFP_HOME/lib
# to compile
cd /vearch/build
mkdir -p /env/app/go/src/github.com/vearch
ln -s /vearch/ /env/app/go/src/github.com/vearch

cd /env/app/go/src/github.com/vearch/vearch/build
./build.sh

mkdir -p /vearch/build/lib/
function copy_deps() {
	# copy from https://h3manth.com/content/copying-shared-library-dependencies
	[[ ! -e $1 ]] && echo "Not a vaild input $1" && exit 1
	[[ -d $2 ]] || echo "No such directory $2 creating..."&& mkdir -p "$2"
	#Get the library dependencies
	echo "Collecting the shared library dependencies for $1..."
	deps=$(ldd $1 | awk 'BEGIN{ORS=" "}$1~/^\//{print $1}$3~/^\//{print $3}' | sed 's/,$/\n/')
	echo "Copying the dependencies to $2"
	#Copy the deps
	for dep in $deps
	do
		echo "Copying $dep to $2"
		cp "$dep" "$2"
	done
	echo "Done!"

}
export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:/vearch/build/gamma_build
copy_deps /vearch/build/bin/vearch /vearch/build/lib
rm -rf /vearch/build/gamma_build
