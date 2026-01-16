#!/bin/bash


go build cmd/kube-simulator/kube-simulator.go

./kube-simulator --reset
