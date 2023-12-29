.DEFAULT_GOAL := build

build:
	export $$(grep -v '^#' .env | xargs) && \
 	go build 
install:
	export $$(grep -v '^#' .env | xargs) && \
 	go install

