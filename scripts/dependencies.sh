#!/bin/sh
set +x

# pip
sudo apt remove -y --purge python-pip && sudo apt-get autoremove
curl https://bootstrap.pypa.io/get-pip.py -o get-pip.py
sudo python ./get-pip.py
sudo -H pip install setuptools -U

# python-dev
sudo apt update && sudo apt install -y python-dev

# sam
sudo apt remove python-six python-yaml python-chardet
sudo -H pip install aws-sam-cli
