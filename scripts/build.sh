#!/bin/bash
rm kube-simulator


go build cmd/kube-simulator/kube-simulator.go

./kube-simulator --reset
