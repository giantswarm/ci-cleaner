#!/bin/sh
set +x

type=${1:-guest}

if [ "$type" = "guest" ]; then
    echo "export AWS_ACCESS_KEY_ID=${GUEST_AWS_ACCESS_KEY_ID}" >> $BASH_ENV
    echo "export AWS_SECRET_ACCESS_KEY=${GUEST_AWS_SECRET_ACCESS_KEY}" >> $BASH_ENV
else
    echo "export AWS_ACCESS_KEY_ID=${HOST_AWS_ACCESS_KEY_ID}" >> $BASH_ENV
    echo "export AWS_SECRET_ACCESS_KEY=${HOST_AWS_SECRET_ACCESS_KEY}" >> $BASH_ENV
fi

. $BASH_ENV
