#!/bin/bash
cd /home/garage/nws_exporter
GOMODCACHE=/home/garage/go/pkg/mod GOCACHE=/home/garage/.cache/go-build go run /home/garage/nws_exporter -station=c5022 -localaddr=0.0.0.0:9883
