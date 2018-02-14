#!/bin/bash

server="127.0.0.1:8080"

curl --data-binary @$1 -H "Content-Type: application/vnd.fk.data+binary" http://$server/messages/ingestion/stream -v
