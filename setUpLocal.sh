#!/bin/bash

#echo "installo netem"
#sudo yum install iproute-tc  #per aws linux
sudo apt-get install iproute2   #per Ubuntu system
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

echo "cambio permessi dei file .sh"
sudo chmod +x simpleNetem.sh
sudo chmod +x variableNetem.sh
sudo chmod +x packetLossTest.sh
sudo chmod +x crashNode.sh
