#! /bin/bash
set -xe

cp -rf conf/coronamans.service /lib/systemd/system/coronamans.service
cp -rf conf/nginx.conf /etc/nginx/nginx.conf
go build .
service coronamans stop
cp -f coronamans /usr/local/bin/
rm -rf /usr/share/nginx/coronamans
cp -rf console/dist /usr/share/nginx/coronamans
service coronamans start
service nginx restart
