#!/bin/bash
export NGINX2SFX_INPUTFILE=/tmp/nginx2sfx.log
export NGINX2SFX_DEBUG=false
export NGINX2SFX_URL=https://ingest.eu0.signalfx.com/v2/datapoint
export NGINX2SFX_TOKEN=your-sfx-token
export NGINX2SFX_INTERVAL=3

./nginx2sfx
