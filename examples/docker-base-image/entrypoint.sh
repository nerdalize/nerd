#!/bin/bash

#Download dataset with nerd command
nerd download $NERD_DATASET_INPUT /in


#Start your own process here
#For example:
touch /out/test.txt


#Upload /out folder
nerd upload /out
