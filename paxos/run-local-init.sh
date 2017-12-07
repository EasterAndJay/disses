#!/bin/bash

CLIENT_COUNT=3

for ((i=5001;i<5001+$CLIENT_COUNT;i+=1))
do
  ttab ./client -cluster-size ${CLIENT_COUNT} -port $i 
done
