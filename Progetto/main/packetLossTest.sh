#!/bin/bash

sudo -v

docker-compose up -d

sudo docker exec node1 tc qdisc add dev eth0 root netem delay 10ms
sudo docker exec node2 tc qdisc add dev eth0 root netem delay 17ms
sudo docker exec node3 tc qdisc add dev eth0 root netem delay 20ms
sudo docker exec node4 tc qdisc add dev eth0 root netem delay 15ms

sleep 12

sudo docker exec node2 tc qdisc add dev eth0 root netem loss 100%

sleep 8

echo "changed loss pack"
sudo docker exec node2 tc qdisc change dev eth0 root netem loss 0%

sleep 10

sudo docker-compose stop
echo "cointainers stopped"

