#!/bin/bash

license_file=./scripts/shell/license.txt

directory=$1

function add_license() {
  find $directory -name '*.py' | while read name
  do
    cat $license_file > /tmp/newfile
    cat $name >> /tmp/newfile
    cp /tmp/newfile $name
  done
}

add_license $directory