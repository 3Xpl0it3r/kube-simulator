#!/bin/bash
rm kube-simulator


go build cmd/kubectl/kubectl.go
go build cmd/kube-simulator/kube-simulator.go

./kube-simulator --reset
