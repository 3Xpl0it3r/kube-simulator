#!/bin/bash


go build cmd/kube-simulator/kube-simulator.go

ip=$(ifconfig|grep 192|awk '{print $2}')
echo ./kube-simulator --cluster-listen ${ip}:6443 --reset
