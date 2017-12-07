IPS=`euca-describe-instances --filter instance-state-name=running | awk '{if (NR % 2 == 0) print $13}'`
PRIVATE_KEY="~/.ssh/cs271-6.pem"

ssh -i $PRIVATE_KEY ubuntu@$IPS[$1] "./paxos-client"

