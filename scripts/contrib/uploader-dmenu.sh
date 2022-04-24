#!/bin/sh

# https://git.tcp.direct/S4D
#    https://git.tcp.direct/tcp.direct/tcp.ac/commit/c18f64059542e35144fd7a479e6758549d017753

function tcpup {
        options="Upload an Image\nRecent Uploads\nCancel"
        selected=$(echo -e $options | dmenu -l 3 )
                if [[ "$selected" == "Upload an Image" ]]; then

                target="$1"
                [ -z "$target" ] && target="$(realpath .)"

ls() {
        echo ..
        command ls -ALNpX1 --group-directories-first "$target"
}

                while :; do
                        sel="$(ls | dmenu -l 25)" || exit
                        if [ "$(echo "$sel")" = "/" ]; then
                                newt="$sel"
                        else
                                newt="$(realpath "$target/$sel")"
                        fi
                        if [ -e "$newt" ]; then
                                target="$newt"
                                if [ ! -d "$target" ]; then
                                        echo "Location: $target"
                                                                command="curl -s -F'upload=@$target' https://tcp.ac/i/put"
                                                echo "Uploading: $target"

                                                url=$(eval $command)
                                                echo "URL: $url"
                                                CLEAN=$(echo $url | sed 's|\\||g' | sed 's|"{"|{"|g' | sed 's|"}"|"}|g')
                                                IMGURL=$(echo $CLEAN | awk -F ':' '{print $2 $3}' | awk -F ',' '{print $1}' | awk -F '}' '{print $1}' | sed -e 's|"||g' -e 's|"||g' -e 's|https|https:|g' -e 's|http\/\/|http:\/\/|g')
                                                echo -n $IMGURL | xclip -sel clip
                                                notify-send "File Uploaded" "URL: $IMGURL \ncopied to clipboard" -t 5000

                                                entry="$(date '+%d-%m-%y-%H:%M:%S')    $url    $(echo $target | awk -F'/' '{print $(NF)}')"
                                                echo $entry >> $HOME/.tcp-uploads
                                                echo "Uploaded"; break
                                fi
                        fi
                done
                elif [[ "$selected" == "Recent Uploads" ]]; then
                    var=$(tac ~/.tcp-uploads | dmenu -l 10)
                    filename=$(echo $var | awk '{print $3}')
                    url=$(echo $var | awk '{print $2}')
                    printf "$url" | xclip -sel clip
                    notify-send "$filename" "URL: $IMGURL \ncopied to clipboard" -t 5000
                elif [[ "$selected" == "Cancel" ]]; then
                                        return
                                fi
}

tcpup
