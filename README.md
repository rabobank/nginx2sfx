### nginx2sfx, read (json) logs from ngins, and send metrics to SignalFx

This cmdline utility reads a json-bases ngins logfile, extracts metrics (requests, response-time, status-code...) from it and sends those as metrics to SignalFx.  
It's primary purpose is to run it as a sidecar process in a Cloud Foundry app container, so we are able to monitor nginx using SignalFx.

## Environment variables
* NGINX2SFX_DEBUG - Put nginx2sfx in debug mode
* NGINX2SFX_INPUTFILE - Name of the file where to read the logs from (default is /tmp/nginx2sfx.log)  
* NGINX2SFX_URL - The URL where to send the metrics (i.e. https://ingest.eu0.signalfx.com/v2/datapoint)   
* NGINX2SFX_SKIP_SSL_VALIDATION - Skip the SSL validation when connecting to SignalFx, default is false
* NGINX2SFX_TOKEN - The SignalFx token to authorize to SignalFx, this value is used in the **X-SF-Token** request header. See below for better security
* NGINX2SFX_BATCH_SIZE - We do not make an http request to sfx for each nginx logline, especially in high volume logging we batch up log lines and send metrics with one http request, default value is 100  
* NGINX2SFX_BATCH_INTERVAL - The maximum time to wait before sending the metrics to SignalFx, default value is 5 seconds  
* NGINX2SFX_URI_AS_DIMENSION - If set to true, the URI will be added as a dimension to the metrics, default is false. (having many different URIs can lead to a lot of dimensions in SignalFx, which will deplete your custom metrics and can be expensive).

## Passing the SignalFx token in a secure way
Although you can use the NGINX2SFX_TOKEN environment variable, it is not recommended to do so for security reasons.  
The better option is to put the token in a credhub service instance (which should have the instance name of _sfxtoken_):  
```
cf cs credhub default sfxtoken -c '{"token":"mySfxToken"}'
```
And then bind this service instance to your app:
```
cf bind-service myApp sfxtoken
``` 
Or specify the service binding in your cf manifest file.

## Metrics

The following json payload will be sent (sample):
```
  {
            "counter": [
                {
                    "metric": "nginx_http_requests_count",
                    "value": 4711,
                    "dimensions": {
                        "uri": "/index.html",
                        "status_code":"200",
                        "cfenv": "d05",
                        "cf_instance_index": 0,
                        "cf_app_name": "my_cf_app",
                        "cf_app_id": "29e3f8f4-f20b-4ca8-b67c-dc72602f7170",
                        "cf_space_name": "abc-space",
                        "cf_org_name": "xyz-org"
                     },
                    "timestamp": 1719998846
                },
                {
                    "metric": "nginx_http_requests_totalTime",
                    "value": 4,
                    "dimensions": {
                        "uri": "/images/whatever.png",
                        "status_code":"404",
                        "cfenv": "d05",
                        "cf_instance_index": 0,
                        "cf_app_name": "my_cf_app",
                        "cf_app_id": "29e3f8f4-f20b-4ca8-b67c-dc72602f7170",
                        "cf_space_name": "abc-space",
                        "cf_org_name": "xyz-org"
                     },
                    "timestamp": 1719998846
                }
            ]
        }
 ```

## How to use

You first need to have a json-based log file from nginx, to get this (in the right format), add the following to your nginx.conf file:
```
log_format sfx escape=json '{"uri":"$uri","method":"$request_method","server_protocol":"$server_protocol","status":"$status","body_bytes_sent":$body_bytes_sent,"request_time":$request_time}';
access_log /tmp/sfx.out sfx;
```

Then you need nginx2sfx to run as a sidecar in your Cloud Foundry app container, to do this add the `sidecars` element to your manifest file, see this example:
```
applications:
- path: .
  name: cf-statics
  memory: 64M
  log-rate-limit-per-second: 25K
  sidecars:
    - name: metrics2sfx
      memory: 32M
      process_types:
      - web
      command: ./nginx2sfx
```

## Download nginx2sfx
The binary can be [downloaded from github](https://github.com/rabobank/nginx2sfx/releases).
