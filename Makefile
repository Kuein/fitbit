.PHONY: build

build:
	sam build

deploy:
	sam build
	sam deploy
