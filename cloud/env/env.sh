#!/usr/bin/env bash

yum update
yum install -y epel-release
yum install -y wget gcc gcc-c++ make automake git blas-devel lapack-devel which openssl-devel libzstd-devel openblas-devel
deploy_home=/env/app
if [ ! -d $deploy_home ]; then
  mkdir -p $deploy_home
fi
zfp_version="0.5.5"
rocksdb_version="v6.2.2"
go_version=1.12.7
cmake_version="3.19.7"
cmake_linux_x86="cmake-${cmake_version}-Linux-x86_64"
cmake_linux_tar_file="${cmake_linux_x86}.tar.gz"
cd $deploy_home
cmake --version
if [ $? -ne 0 ]; then
    wget https://github.com/Kitware/CMake/releases/download/v${cmake_version}/${cmake_linux_tar_file}
    tar xf $cmake_linux_tar_file -C $deploy_home
    add_sys_path="export PATH=$deploy_home/${cmake_linux_x86}/bin:$PATH"
    export_cmake_root="export CMAKE_ROOT=$deploy_home/${cmake_linux_x86}"
    $add_sys_path
    $export_cmake_root
    echo "$export_cmake_root $add_sys_path"
    echo -e "$make_path\n$export_cmake_root\n" >> "/root/.bashrc"
fi
if [ ! -d "zfp" ]; then
    wget https://github.com/LLNL/zfp/archive/${zfp_version}.tar.gz -O zfp.tar.gz
    tar -xzvf zfp.tar.gz
    cd $deploy_home/zfp-${zfp_version}
    mkdir build && cd build
    cmake -DCMAKE_INSTALL_PREFIX=$deploy_home/zfp_install ..
    cmake --build --config Release .
    make install
fi
cd /env/app
if [ ! -f "rocksdb-${rocksdb_version}.tar.gz" ]; then
    wget https://github.com/facebook/rocksdb/archive/${rocksdb_version}.tar.gz -O rocksdb-${rocksdb_version}.tar.gz
fi
tar -xzf rocksdb-v${rocksdb_version}.tar.gz
cd $deploy_home/rocksdb-${rocksdb_version}
make shared_lib -j30
make install-shared INSTALL_PATH=$deploy_home/rocksdb_install
cd /$deploy_home
if [ ! -f "go${go_version}.linux-amd64.tar.gz" ]; then
    wget https://dl.google.com/go/go${go_version}.linux-amd64.tar.gz
fi
tar -xzf go${go_version}.linux-amd64.tar.gz
#rm *tar.gz
