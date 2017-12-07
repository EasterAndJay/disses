#!/bin/bash

CLIENT_COUNT=5

for ((i=3;i<$CLIENT_COUNT;i+=1))
do
  ttab ./client -cluster-size ${CLIENT_COUNT} -id $i
done
