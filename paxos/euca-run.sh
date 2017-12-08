IFS=' ' read -r -a IPS <<< "`euca-describe-instances --filter instance-state-name=running | awk '{if (NR % 2 == 0) print $13}'`"

PRIVATE_KEY="~/.ssh/cs271-6.pem"
CLUSTER_SIZE=0

if [ $1 -gt 2 ]
then
  CLUSTER_SIZE=5
else
  CLUSTER_SIZE=3
fi
echo ${IPS[$1]}
# ssh -i $PRIVATE_KEY ubuntu@$IPS[$1] -t "./client -cluster-size $CLUSTER_SIZE -id $1 &" >> "client$1.log"
ssh -i $PRIVATE_KEY ubuntu@169.231.235.9 -t "./client -cluster-size $CLUSTER_SIZE -id $1 &" >> "client$1.log"


