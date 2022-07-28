#!/bin/env bash
# Author: Perp
# Description: Upload screenshots to tcp.ac (rofi)
# Usage: upload

# Screenshot tool
SCREENSHOT="flameshot gui"
# Image directory
DIRECTORY="$HOME/Pictures"

# Choose mode for screenshot or upload
function mode_1() {
    case "$1" in
        "") sleep 1 && $($SCREENSHOT) ;;
        "") mode_2 $(echo -e " Clipboard\n File" | rofi -dmenu -i) ;;
        *) exit 1 ;;
    esac
}

# Choose mode for clipboard or file upload
function mode_2() {
    case "$1" in
        "") choosen "clipboard" ;;
        "") choosen "file" ;;
        *) exit 1 ;;
    esac
}

# Check chosen upload option
function choosen() {
    case "$1" in
        "clipboard")
            # Save image to temp file
            xclip -sel clip -t image/png -o > /tmp/image.png
            # Upload to tcp.ac
            upload "/tmp/image.png"
            # Delete image
            rm /tmp/image.png
            ;;
        "file")
            # Choose a image file
            FILE=$(echo -e "$(find $DIRECTORY -type f \( -name "*.png" -o -name "*.jpg" -o -name "*.jpeg" \))" | rofi -dmenu -i)
            # Upload to tcp.ac
            upload $FILE
            ;;
        *) exit 1 ;;
    esac
}

# Upload to tcp.ac
function upload() {
    # Upload to tcp.ac
    OUT=$(curl -s -F "upload=@$1" https://tcp.ac/i/put)
    # Check for an upload error
    error_check $OUT
    # Parse the json
    OUT=$(echo $OUT | jq ".Imgurl, .ToDelete" -r)
    # Convert spaces to newline
    URLS=$(echo $OUT | sed 's/\s\+/\n/g')
    # Display a notification
    notify-send -t 10000 "Upload to tcp.ac success (Check clipboard)"
    # Write to log file
    echo $URLS | sed 's/ /:/g' >> /tmp/tcp.ac
    # Copy upload link to clipboard
    echo -e "$URLS" | head -1 | xclip -sel clip
}

# Check a upload error
function error_check() {
    # Image URL not found
    if ! [[ "$1" == *"Imgurl"* ]];
    then
        # Display a notification
        notify-send -t 10000 "Upload to tcp.ac failed ($1)"
        exit 1
    fi
}

# Start the mode for screenshot or upload
mode_1 $(echo -e "  Screenshot\n  Upload" | rofi -dmenu -i)
