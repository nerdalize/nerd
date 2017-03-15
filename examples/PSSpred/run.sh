#!/bin/bash

# if [ -z "$1" ]
#   then echo "Usage: $0 <inputfile.ext>"
#   exit 0
# fi




nerd download $NERD_DATASET_INPUT /in/

echo "starting psspred with inputfile"

/library/PSSpred/PSSpred.pl /in/sequence.fa

cat "/library/PSSpred/seq.dat.ss" |  grep -v "coil  helix  beta" | awk -v OFS="\n" '{if($3 == "C") print "L:"$4;if($3 == "H") print "H:"$5; if($3 == "E") print "E:"$6;}' | tr "\n" "," > /out/output.txt

echo "process complete"