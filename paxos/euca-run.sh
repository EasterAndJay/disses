IPS=`euca-describe-instances --filter instance-state-name=running | awk '{if (NR % 2 == 0) print $13}'`
PRIVATE_KEY="~/.ssh/cs271-6.pem"
CLUSTER_SIZE=0

if [ $1 -gt 2 ]
then
  CLUSTER_SIZE=5
else
  CLUSTER_SIZE=3
fi

ssh -i $PRIVATE_KEY ubuntu@$IPS[$1] "./paxos-client -cluster-size $CLUSTER_SIZE -id $1"

