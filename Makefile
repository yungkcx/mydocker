mydocker = ~/Golang/bin/mydocker
image = abc
# command = stress --vm-bytes 200m --vm-keep -m 1
# command = sleep 30
command = bash

all: build

.PHONY: build run cmd
build:
	go install

run: build
	sudo $(mydocker) run --ti -v ~/git/docker/cmnt:/mnt -m 100m $(image) $(command)
	# sudo $(mydocker) run -d -m 100m $(image) $(command)

cmd: build
	# sudo $(mydocker) rmi a b c d e
	sudo $(mydocker) ps
	# sudo $(mydocker) image ubuntu.tar -o xx