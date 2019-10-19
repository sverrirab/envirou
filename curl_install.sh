#!/bin/bash
set -e

function clean_dir() {
    if [ -d $1 ] ; then
        rm -rf $1.old
        mv $1 $1.old
    fi
    mkdir -p $1
}

cd
clean_dir .envirou

echo "Downloading Envirou..."
cd .envirou
curl -S -s -L -o- https://github.com/sverrirab/envirou/archive/master.tar.gz | tar zx

echo ""
echo "Installing..."
./envirou-master/install

cd
rm -rf .envirou.old
