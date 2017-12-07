#!/bin/bash

euca-run-instances emi-ff557930 -n 5 -g default -k cs271-6 -t m3.2xlarge -z race

euca-authorize default -P tcp -p 22 -s 0.0.0.0/0
euca-authorize default -P udp -p 5000 -s 0.0.0.0/0

