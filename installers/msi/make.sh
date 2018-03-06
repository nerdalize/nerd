#! /bin/bash
set -e

cp $GOPATH/bin/nerd ./nerd.exe
VERSION=$(cat ../../VERSION | cut -f1 -d"-")

sed -i "s/\sVersion=".*"/ Version=\"$VERSION\"/" Product.wxs

$WINDIR/Microsoft.NET/Framework/v4.0.30319/MSBuild.exe nerd.wixproject
