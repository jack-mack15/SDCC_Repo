#!/bin/bash

sudo -v

docker-compose up -d

sudo docker exec node1 tc qdisc add dev eth0 root netem delay 10ms
sudo docker exec node2 tc qdisc add dev eth0 root netem delay 17ms
sudo docker exec node3 tc qdisc add dev eth0 root netem delay 20ms
sudo docker exec node4 tc qdisc add dev eth0 root netem delay 15ms

echo "attesa"
sleep 10

echo "killing node2"
sudo docker kill node2

sleep 10

sudo docker-compose stop
echo "containers stopped"