PROJECTNAME := $(shell basename "$(PWD)")
include .env
export $(shell sed 's/=.*//' .env)

# Dockerfile
## gen-images: Generate serivces' image
.PHONY: gen-images
gen-images:
	@docker build --tag bank-svc:$(shell git rev-parse --short HEAD) -f ./build/Dockerfile .

## run-service:
.PHONY: 
run-service:
	@docker run --env-file ./.env -v ./deployment/application.yaml:/application.yaml -p 8080:8080 bank-svc:$(shell git rev-parse --short HEAD) 