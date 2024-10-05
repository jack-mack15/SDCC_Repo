#!/bin/bash

sudo -v

docker-compose up -d

sudo docker exec node1 tc qdisc add dev eth0 root netem delay 20ms 5ms
sudo docker exec node2 tc qdisc add dev eth0 root netem delay 34ms 7ms
sudo docker exec node3 tc qdisc add dev eth0 root netem delay 40ms 8ms
sudo docker exec node4 tc qdisc add dev eth0 root netem delay 30ms 4ms
sudo docker exec node5 tc qdisc add dev eth0 root netem delay 20ms 3ms

sleep 60

sudo docker-compose stop
echo "containers stopped"