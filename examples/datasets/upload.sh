#!/bin/bash
aws s3 cp helloworld.zip s3://nlz-datasets-dev/test-helloworld.zip --grants read=uri=http://acs.amazonaws.com/groups/global/AllUsers
