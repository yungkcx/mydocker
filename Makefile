mydocker = ~/Golang/bin/mydocker
image = abc
command = bash

all: build

.PHONY: build run cmd
build:
	go install

run: build
	sudo $(mydocker) run --ti -v ~/git/docker/cmnt:/mnt -m 100m $(image) $(command)

cmd: build
	# sudo $(mydocker) rmi a b c d
	sudo $(mydocker) images
	# sudo $(mydocker) image ubuntu.tar -o xx