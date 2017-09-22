all: build

.PHONY: build run
build:
	go install

run: build
	sudo ~/Golang/bin/mydocker run -ti -v ~/git/docker/cmnt:/ sh