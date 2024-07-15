#!/bin/bash
#
while true
do
  curl -sk https://cf-statics.apps.cfd05.rabobank.nl/ -o /dev/null
  curl -sk https://cf-statics.apps.cfd05.rabobank.nl/dir-index-test/KeepAlive-test/ff01.jpg -o /dev/null
  curl -sk 'https://cf-statics.apps.cfd05.rabobank.nl/dir-index-test/KeepAlive-test/ff02.jpg?parm1=whatever' -o /dev/null
  curl -sk --http1.0 https://cf-statics.apps.cfd05.rabobank.nl -o /dev/null
  curl -sk https://cf-statics.apps.cfd05.rabobank.nl/ -o /dev/null
  curl -sk https://cf-statics.apps.cfd05.rabobank.nl/dir-index-test/KeepAlive-test/iserniet -o /dev/null
  curl -sk https://cf-statics.apps.cfd05.rabobank.nl/dir-index-test/KeepAlive-test/ff01.jpg -o /dev/null
  curl -sk 'https://cf-statics.apps.cfd05.rabobank.nl/dir-index-test/KeepAlive-test/ff02.jpg?parm1=whatever' -o /dev/null
  curl -sk --http1.0 https://cf-statics.apps.cfd05.rabobank.nl -o /dev/null
  sleep 0.1
done
