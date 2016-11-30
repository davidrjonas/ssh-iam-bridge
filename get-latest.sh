#!/bin/sh -e

BIN=/usr/sbin/ssh-iam-bridge
URL=https://github.com/davidrjonas/ssh-iam-bridge

LATEST=$(curl -sL -H 'Accept: application/json' $URL/releases/latest | perl -ne 'print "$1\n" if /tag_name":"([\d\.]+)"/')

echo Downloading version $LATEST

curl -sL $URL/releases/download/${LATEST}/ssh-iam-bridge > $BIN
chmod 0755 $BIN

echo Installed as $BIN
