#!/bin/sh
# count = 1
docker build . -t proj2-test

while IFS= read -r line
do
echo "$line"
docker run --name $line --network mynetwork -h $line proj2-test -h hostfile &
sleep 1s
done < hostfile

sleep 10s
while IFS= read -r line
do
docker stop $line
docker logs $line >& ./logs/$line.log
docker rm $line
done < hostfile
echo " "
echo "Done"


