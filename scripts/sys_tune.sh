#!/bin/sh

sysctl -w fs.file-max=12000000
 
sysctl -w fs.nr_open=11000000
 
ulimit -n 3000000
 
sysctl -w net.ipv4.ip_local_port_range="500  65535"
 
sysctl -w net.ipv4.tcp_rmem="1024 4096 16384"
sysctl -w net.ipv4.tcp_wmem="1024 4096 16384"
sysctl -w net.ipv4.tcp_moderate_rcvbuf="0"
 
sysctl -w net.ipv4.tcp_tw_recycle=1
