#!/bin/bash

CLIENT_COUNT=3

for ((i=0;i<$CLIENT_COUNT;i+=1))
do
  ttab ./client -cluster-size ${CLIENT_COUNT} -id $i 
done
