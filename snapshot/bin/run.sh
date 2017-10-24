#!/bin/bash

CLIENT_COUNT=2

for ((i=0;i<$CLIENT_COUNT;i+=1))
do
  ttab bin/snapshot -n ${CLIENT_COUNT} -p ${i}
done