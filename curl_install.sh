#!/bin/bash
set -e
INSTALL_DIR=".envirou"
upgrade_only=0

function clean_dir() {
    if [ -d $1 ] ; then
        rm -rf $1.old
        mv $1 $1.old
    fi
    mkdir -p $1
}

cd
if [ -d "${INSTALL_DIR}" ] ; then
    upgrade_only=1
fi
clean_dir "${INSTALL_DIR}"

cd "${INSTALL_DIR}"
echo "Downloading Envirou..."
curl -S -s -L -o- https://github.com/sverrirab/envirou/archive/master.tar.gz | tar zx
cd

if [ ${upgrade_only} -eq 0 ] ; then
    echo ""
    echo "Installing..."
    "${INSTALL_DIR}/envirou-master/install"
else
    echo "Upgraded successfully. If Envirou does not work try runnning install again:"
    echo "~/${INSTALL_DIR}/envirou-master/install"
fi

rm -rf .envirou.old
