#!/bin/sh
set -o errexit
set -o pipefail

VENDOR=nerdalize.com
DRIVER=dataset

driver_dir=$VENDOR${VENDOR:+"~"}${DRIVER}
if [ ! -d "/flexmnt/$driver_dir" ]; then
  mkdir "/flexmnt/$driver_dir"
fi

# atomically write (new) flex volume executable
cp "/$DRIVER" "/flexmnt/$driver_dir/.$DRIVER"
mv -f "/flexmnt/$driver_dir/.$DRIVER" "/flexmnt/$driver_dir/$DRIVER"

# copy service account from pod to the host (for the flex volume)
cp -R /var/run/secrets/kubernetes.io/serviceaccount /flexmnt/$driver_dir

# write env information (for api host info and such)
env > /flexmnt/$driver_dir/flex.env

# block forever to keep the daemonset running
while : ; do
  sleep 3600
done
