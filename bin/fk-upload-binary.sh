#!/bin/bash

curl --data-binary @$1 -H "Content-Type: application/vnd.fk.data+binary" http://127.0.0.1:8080/messages/ingestion/stream -v
