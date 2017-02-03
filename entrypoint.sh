#!/bin/sh
mkdir /s3data

echo "Downloading"
nerd download $DATASET /s3data

touch /s3data/en.txt
touch /s3data/pils.txt

echo "Uploading"
nerd upload $DATASET /s3data
