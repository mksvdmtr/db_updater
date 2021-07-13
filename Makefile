.PHONY: build

bin_name=db_updater

build:
		go build -o $(bin_name) *.go

.DEFAULT_GOAL := build
