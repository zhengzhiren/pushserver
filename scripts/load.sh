#!/bin/bash

SCRIPT=$0
DEVICE_COUNT=1
APP_COUNT=1
LOOP=0

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
  echo "${REV}-f${NORM}  --Start and kill the applications frequently."
  echo -e "${REV}-h${NORM}  --Displays this help message. No further functions are performed."\\n
  echo -e "Example: ${BOLD}$SCRIPT -a 5 -d 10 127.0.0.1 9999${NORM}"\\n
  echo -e "Example: ${BOLD}$SCRIPT 10.135.28.70 20000 10.135.28.72 20000${NORM}"\\n
  exit 1
}

function RUN {
	i=1
	while [ $i -le $DEVICE_COUNT ]; do
		TEMP_PORT=$(($SDK_PORT+$i))
		DEV_ID=$DEVICE_ID$i
		# randomly choose a push server
		index=$(($RANDOM%$SERVER_COUNT))
		PUSHIP=${IPPORT[${index}*2]}
		PUSHPORT=${IPPORT[${index}*2+1]}
		echo "Device [$DEV_ID] is connecting $PUSHIP:$PUSHPORT"
		simsdk -i "$DEV_ID" -p $TEMP_PORT $PUSHIP:$PUSHPORT > "$DEV_ID.out"&
		let i=i+1
	done

	sleep 5  #sleep a while to let the simsdk get ready
	i=1
	while [ $i -le $DEVICE_COUNT ]; do
		TEMP_PORT=$(($SDK_PORT+$i))
		DEV_ID=$DEVICE_ID$i
		j=1
		while [ $j -le $APP_COUNT ]; do
			APP_ID="testapp$j"
			echo "Starting APP [$APP_ID] on Device [$DEV_ID]"
			simapp -p $TEMP_PORT -r $(($APP_PORT+$i*$APP_COUNT+$j)) "$APP_ID" "AppKey_$j" > "$DEV_ID-$APP_ID.out" &
			let j=j+1
		done
		let i=i+1
	done
}

STOP()
{
	killall simsdk simapp
}

BASHTRAP()
{
	echo "CTRL+C Detected! Stopping..."
	STOP
	exit
}

#process the command arguments
while getopts "d:a:fh" arg
do
	case $arg in
		a)
		APP_COUNT=$OPTARG
		;;
		d)
		DEVICE_COUNT=$OPTARG
		;;
		f)
		LOOP=1
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

if [ $# -eq 0 -o $(($#%2)) -ne 0 ]; then
	HELP
fi

IPPORT=("$@")
SERVER_COUNT=$((${#IPPORT[@]}/2))

#echo $@
#echo ${IPPORT[@]}
echo "$SERVER_COUNT servers: ${IPPORT[@]}"

trap BASHTRAP INT

if [ $LOOP -eq 1 ]; then
	while :
	do
		RUN
		sleep 2
		STOP
	done
else
	RUN
	echo "Press CTRL+C to stop all Apps"
	sleep 3600
fi
