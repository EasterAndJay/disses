#!/bin/bash

IPS=`euca-describe-instances --filter instance-state-name=running | awk '{if (NR % 2 == 0) print $13":"$8+5001}'`
PRIVATE_KEY="~/.ssh/cs271-6.pem"
EXECUTABLE="client"
PEERS_FILE="peersRemote.txt"
echo "ip-port pairs are:"
echo $IPS
make euca

for ip_port in $IPS
do
  ip=`echo $ip_port | cut -d ':' -f 1`
  host="ubuntu@$ip"
  # echo $ip_port2 >> "$PEERS_FILE"
  # echo -e "`echo $ip_port | cut -d ':' -f 2`\n$(cat $PEERS_FILE)" > $PEERS_FILE
  # echo "peers.txt file for host: $ip_port"
  # cat $PEERS_FILE
  scp -i $PRIVATE_KEY $EXECUTABLE "$host:~/$EXECUTABLE"
  scp -i $PRIVATE_KEY $PEERS_FILE "$host:~/peers.txt"
  # rm $PEERS_FILE 
done
