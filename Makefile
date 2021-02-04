ifneq (,$(wildcard ./.env.override))
    include .env.override
    export
endif

DOCKER_COMPOSE = ./bin/docker-compose

define COMPOSE_TEMPLATE
version: '3.7'
services:
  mongo:
    image: mongo:_VERSION_
endef

export $COMPOSE_TEMPLATE

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
start: setup upgrade_mongodb
ifeq ($(SHELLHUB_ENV),development)
	@echo Starting ShellHub in development mode...
	@echo
	@$(DOCKER_COMPOSE) up
else
	@echo Starting ShellHub in production mode...
	@echo

	@$(DOCKER_COMPOSE) up -d
endif

.PHONY: stop
## Stop services
stop:
	@$(DOCKER_COMPOSE) stop

.PHONY: build
## Build all services (append "SERVICE=<service>" to build a specific one)
build: check_development_mode
	@$(DOCKER_COMPOSE) build $(SERVICE)

.PHONY: check_development
# Check for development mode
check_development_mode:
ifneq ($(SHELLHUB_ENV),development)
	@echo Development mode disabled!
	@exit 1
endif

.SILENT:
.ONESHELL:
upgrade_mongodb: SHELL := sh
upgrade_mongodb:
	# Wait for mongodb to be available
	wait_for_mongo() {
		# Command used to ping mongodb
		MONGO_PING_CMD=$$(cat <<-EOF
			db.runCommand({ ping: 1 })
		EOF
		)

		while ! $(DOCKER_COMPOSE) exec mongo mongo --quiet --eval "quit($$MONGO_PING_CMD ? 0 : 1)" >/dev/null 2>&1; do
			sleep 1
		done
	}

	# Upgrade mongodb version to $1
	upgrade_mongo() {
		export EXTRA_COMPOSE_FILE=$$(mktemp)
		echo "$$COMPOSE_TEMPLATE" | sed "s,_VERSION_,$$1,g" > $$EXTRA_COMPOSE_FILE

		$(DOCKER_COMPOSE) stop mongo
		$(DOCKER_COMPOSE) up -d mongo

		wait_for_mongo

		# Command used to set compatibility version
		MONGO_SET_COMPAT_VERSION_CMD=$$(cat <<-EOF
			db.adminCommand({
				setFeatureCompatibilityVersion: '$$1'
			})
		EOF
		)

		$(DOCKER_COMPOSE) exec mongo mongo --quiet --eval "quit($${MONGO_SET_COMPAT_VERSION_CMD}.ok ? 1 : 0)"
	}

	$(DOCKER_COMPOSE) up -d mongo

	wait_for_mongo

	# Command used to get compatibility version
	MONGO_GET_COMPAT_VERSION_CMD=$$(cat <<-EOF
		db.adminCommand({
			getParameter: 1,
			featureCompatibilityVersion: 1
		})['featureCompatibilityVersion'].version
	EOF
	)

	# Before upgrading to 4.2-series, we need to upgrade earlier versions first
	while true; do
		MONGO_COMPAT_VERSION=$$($(DOCKER_COMPOSE) exec mongo mongo --quiet --eval "$$MONGO_GET_COMPAT_VERSION_CMD" | tr -d '\r')
		case $$MONGO_COMPAT_VERSION in
			# Upgrade from 3.4 to 3.6
			3.4*) upgrade_mongo 3.6.21 ;;
			# Upgrade from 3.6 to 4.0
			3.6*) upgrade_mongo 4.0.22 ;;
			# Upgrade from 4.0 to 4.2
			4.0*) upgrade_mongo 4.2.12 ;;
			# Latest version
			4.2*) break ;;
	      esac
	done

.PHONY: help
help:
	@echo "$$(tput bold)Available commands:$$(tput sgr0)";echo;sed -ne"/^## /{h;s/.*//;:d" -e"H;n;s/^## //;td" -e"s/:.*//;G;s/\\n## /---/;s/\\n/ /g;p;}" ${MAKEFILE_LIST}|LC_ALL='C' sort -f|awk -F --- -v n=$$(tput cols) -v i=19 -v a="$$(tput setaf 6)" -v z="$$(tput sgr0)" '{printf"%s%*s%s ",a,-i,$$1,z;m=split($$2,w," ");l=n-i;for(j=1;j<=m;j++){l-=length(w[j])+1;if(l<= 0){l=n-i-length(w[j])-1;printf"\n%*s ",-i," ";}printf"%s ",w[j];}printf"\n";}'|more $(shell test $(shell uname) == Darwin && echo '-Xr')

.DEFAULT_GOAL := help
