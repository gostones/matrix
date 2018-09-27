#!/usr/bin/env bash

#args:

#DRY_RUN=echo

name=$1
[[ $# -lt 1 ]] && name="matrix"

##
function deploy() {
    local domain="run.aws-usw02-pr.ice.predix.io"

    echo "Pushing service $domain $name ..."

    #$DRY_RUN cf delete-route $domain --hostname $name -f

    $DRY_RUN cf push -f manifest.yml -d $domain --hostname $name; if [ $? -ne 0 ]; then
        return 1
    else
        return 0
    fi
}

#
echo "### Deploying ..."

deploy; if [ $? -ne 0 ]; then
    echo "#### Deploy failed"
    exit 1
fi

exit 0
##
