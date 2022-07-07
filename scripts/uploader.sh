#!/usr/bin/env bash

#
# tcp.ac uploader script, works in terminal or can be used with scrot.
#
# example openbox configuration to use scrot on the printscreen keybinding for easy image uploads:
#
#    <keybind key="Print">
#      <action name="Execute">
#        <command>scrot -f -s -e 'tcpac $f'</command>
#      </action>
#    </keybind>
#

####################### https://github.com/tlatsas/bash-spinner

function _spinner() {
    # $1 start/stop
    #
    # on start: $2 display message
    # on stop : $2 process exit status
    #           $3 spinner function pid (supplied from stop_spinner)

    local on_success="DONE"
    local on_fail="FAIL"
    local green="\e[1;32m"
    local red="\e[1;31m"
    local nc="\e[0m"

    case $1 in
        start)
            # calculate the column where spinner and status msg will be displayed
            let column=$(tput cols)-${#2}-8
            # display message and position the cursor in $column column
            echo -ne ${2}
            printf "%${column}s"

            # start spinner
            i=1
            sp='\|/-'
            delay=${SPINNER_DELAY:-0.15}

            while :
            do
                printf "\b${sp:i++%${#sp}:1}"
                sleep $delay
            done
            ;;
        stop)
            if [[ -z ${3} ]]; then
                echo "spinner is not running.."
                exit 1
            fi

            kill $3 > /dev/null 2>&1

            # inform the user uppon success or failure
            echo -en "\b["
            if [[ $2 -eq 0 ]]; then
                echo -en "${green}${on_success}${nc}"
            else
                echo -en "${red}${on_fail}${nc}"
            fi
            echo -e "]"
            ;;
        *)
            echo "invalid argument, try {start/stop}"
            exit 1
            ;;
    esac
}

function start_spinner {
    # $1 : msg to display
    _spinner "start" "${1}" &
    # set global spinner pid
    _sp_pid=$!
    disown
}

function stop_spinner {
    # $1 : command exit status
    _spinner "stop" $1 $_sp_pid
    unset _sp_pid
}

################################################
start_spinner "uploading $1..."
OUT=$(curl -s -F "upload=@$1" https://tcp.ac/i/put)
if ! [[ "$OUT" == *"Imgurl"* ]]; then
	echo ""
	echo "ERROR: $OUT"
	notify-send -i network-error 'tcp.ac upload failed' "$OUT" -t 5000
	stop_spinner 2
	exit 2
else
	_CLEAN=$(echo "$OUT" | sed 's|\\||g' | sed 's|"{"|{"|g' | sed 's|"}"|"}|g')
	echo "$_CLEAN" | jq | tee "$1.DELETEKEY"
	_IMGURL=$(echo "$_CLEAN" | awk -F ':' '{print $2 $3}' | awk -F ',' '{print $1}' | awk -F '}' '{print $1}' | sed -e 's|"||g' -e 's|"||g' -e 's|https|https:|g' -e 's|http\/\/|http:\/\/|g')
	echo -n "$_IMGURL" | xclip -sel clip
	notify-send -i image-x-generic 'tcp.ac upload success' 'link copied to clipboard, delete key saved in file adjacent to original image' -t 5000
fi
stop_spinner $?
