#!/bin/bash

sudo -v

docker-compose up -d

sudo docker exec node1 tc qdisc add dev eth0 root netem delay 10ms
sudo docker exec node2 tc qdisc add dev eth0 root netem delay 17ms
sudo docker exec node3 tc qdisc add dev eth0 root netem delay 20ms
sudo docker exec node4 tc qdisc add dev eth0 root netem delay 15ms
sudo docker exec node5 tc qdisc add dev eth0 root netem delay 10ms

sleep 60

sudo docker-compose stop
echo "containers stopped"
