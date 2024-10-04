#!/bin/bash

echo "installo netem"
sudo yum install iproute-tc
echo "carico netem nel kernel"
sudo modprobe sch_netem

echo "build image dei nodi"
cd node
sudo docker build -f Dockerfile.node -t node .
cd ..

echo "build image del registry"
cd registry
sudo docker build -f Dockerfile.registry -t registry .
cd ..

echo "cambio diritti dei file .sh"
sudo chmod +x simpleNetem.sh
sudo chmod +x variableNetem.sh
sudo chmod +x packetLossTest.sh
