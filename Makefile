ifneq (,$(wildcard ./.env.override))
    include .env.override
    export
endif

# Generate required private key for api service
api_private_key:
	@openssl genrsa -out api_private_key 2048

# Generate required public key for api service
api_public_key:
	@openssl rsa -in api_private_key -out api_public_key -pubout

# Generate required private key for ssh service
ssh_private_key:
	@openssl genrsa -out ssh_private_key 2048

.PHONY: setup
# Setup required files
setup: api_private_key api_public_key ssh_private_key

.PHONY: start
## Start services
start: setup
ifeq ($(SHELLHUB_ENV),development)
	@echo Starting ShellHub in development mode...
	@echo
	@./bin/docker-compose up
else
	@echo Starting ShellHub in production mode...
	@echo
	@./bin/docker-compose up -d
endif

.PHONY: stop
## Stop services
stop:
	@./bin/docker-compose stop

.PHONY: build
## Build all services (append "SERVICE=<service>" to build a specific one)
build: check_development_mode
	@./bin/docker-compose build $(SERVICE)

.PHONY: check_development
# Check for development mode
check_development_mode:
ifneq ($(SHELLHUB_ENV),development)
	@echo Development mode disabled!
	@exit 1
endif

.PHONY: help
help:
	@echo "$$(tput bold)Available commands:$$(tput sgr0)";echo;sed -ne"/^## /{h;s/.*//;:d" -e"H;n;s/^## //;td" -e"s/:.*//;G;s/\\n## /---/;s/\\n/ /g;p;}" ${MAKEFILE_LIST}|LC_ALL='C' sort -f|awk -F --- -v n=$$(tput cols) -v i=19 -v a="$$(tput setaf 6)" -v z="$$(tput sgr0)" '{printf"%s%*s%s ",a,-i,$$1,z;m=split($$2,w," ");l=n-i;for(j=1;j<=m;j++){l-=length(w[j])+1;if(l<= 0){l=n-i-length(w[j])-1;printf"\n%*s ",-i," ";}printf"%s ",w[j];}printf"\n";}'|more $(shell test $(shell uname) == Darwin && echo '-Xr')

.DEFAULT_GOAL := help
