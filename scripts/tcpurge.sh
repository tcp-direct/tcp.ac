#!/usr/bin/env bash
echo "Deleting all remote copies of tcp.ac images made by uploader.sh in $(pwd)/...."
echo "sleeping for 5 seconds, CTRL+C to cancel"
sleep 5
find "$(pwd)/" -maxdepth 1 -iname "*.DELETEKEY" -print | while read -r line; do
	echo "$line"
	cat "$line" | \
	grep "ToDelete" | \
	awk '{print $2}' | tr -d '"' | \
	while read -r line; do
		curl -s "$line" & sleep 0.5;
	done
done;

