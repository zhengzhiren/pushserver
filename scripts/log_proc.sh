#!/bin/bash

#This script process the log files to check if we received good messages

#fonts
NORM=`tput sgr0`
REV=`tput smso`

file=""
for f in $( ls device_*-*.out ); do
	# filter the received messages out
	grep "Received message." $f | awk '{ $1=""; print $0 }' > "$f.msg"
	if [ $file="" ]; then
		file="$f.msg"
	fi
done 

PASS=1
for f in $( ls device_*-*.out.msg ); do
	if [ ! -s $f ]; then
		echo "Error! ${REV}$f${NORM} is empty"
	else
		echo "Comparing \"$file\" with \"$f\"..."
		cmp -s $file $f
		if [ $? -ne 0 ]; then
			echo "Error! Different file content: ${REV}\"$file\"${NORM} ${REV}\"$f\"${NORM}"
			PASS=0
		fi
	fi
done

if [ $PASS -eq 1 ]; then
	echo "All messages are OK!"
else
	echo "The test fails!"
fi
