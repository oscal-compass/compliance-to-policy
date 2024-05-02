#!/bin/sh

version="v1.27.3"

arch=`uname -m`
if [ "$arch" == "x86_64" ];then
  arch="amd64"
elif [ "$arch" == "aarch64" ];then
  arch="arm64"
else
  arch="amd64"
fi

wget https://dl.k8s.io/release/$version/bin/linux/$arch/kubectl
chmod +x kubectl
mv kubectl /usr/local/bin/