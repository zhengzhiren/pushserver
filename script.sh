#!/bin/bash

SCRIPT=$0
DEVICE_COUNT=5
APP_COUNT=3

DEVICE_ID="device_"
SDK_PORT=13000
APP_PORT=31000

#Set fonts for Help.
NORM=`tput sgr0`
BOLD=`tput bold`
REV=`tput smso`

#Help function
function HELP {
  echo -e "${REV}Usage:${NORM} ${BOLD}$SCRIPT [option] <ip> <port>${NORM}"\\n
  echo "Options:"
  echo "${REV}-a${NORM}  --Sets the ${BOLD}App count on each device${NORM}. Default is ${BOLD}3${NORM}."
  echo "${REV}-d${NORM}  --Sets the ${BOLD}device count${NORM}. Default is ${BOLD}5${NORM}."
  echo -e "${REV}-h${NORM}  --Displays this help message. No further functions are performed."\\n
  echo -e "Example: ${BOLD}$SCRIPT -a 5 -d 10 127.0.0.1 9999${NORM}"\\n
  exit 1
}

#process the command arguments
while getopts "d:a:h" arg
do
	case $arg in
		a)
		APP_COUNT=$OPTARG
		;;
		d)
		DEVICE_COUNT=$OPTARG
		;;
		h)
		HELP
		;;
		?)
		echo "unkonw argument"
		HELP
		;;
	esac
done

shift $((OPTIND-1))  #This tells getopts to move on to the next argument.

if [ $# -ne 2 ]; then
	HELP
fi

IP=$1
PORT=$2

i=0
while [ $i -lt $DEVICE_COUNT ]; do
	TEMP_PORT=$(($SDK_PORT+$i))
	DEV_ID=$DEVICE_ID$i
	echo "Starting Device [$DEV_ID]"
	simsdk -i "$DEV_ID" -p $TEMP_PORT $IP:$PORT > "$DEV_ID.out"&
	j=0
	while [ $j -lt $APP_COUNT ]; do
		APP_ID="app_$j"
		echo "Starting APP [$APP_ID] on Device [$DEV_ID]"
		simapp -p $TEMP_PORT -r $(($APP_PORT+$i*$APP_COUNT+$j)) "$APP_ID" "AppKey_$j" > "$DEV_ID-$APP_ID.out" &
		let j=j+1
	done
	let i=i+1
done
