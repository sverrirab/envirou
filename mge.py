#!/bin/bash

#export MGE_DEBUG="$ZSH_VERSION / $BASH_VERSION"

if [ -n "${ZSH_VERSION}" ] ; then
    # Support source-ing from zsh
    MGE_ROOT=$(dirname ${(%):-%x})
else
    MGE_ROOT=$(dirname "${BASH_SOURCE[0]}")
fi

$(python "${MGE_ROOT}/mge.py" $*)
