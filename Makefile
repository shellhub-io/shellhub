ENV_OVERRIDE ?= ./.env.override

ifneq (,$(wildcard $(ENV_OVERRIDE)))
    include $(ENV_OVERRIDE)
endif

DOCKER_COMPOSE = ./bin/docker-compose
KEYGEN = ./bin/keygen

# Generate required private key for api service
api_private_key:
	@$(KEYGEN) genpkey -algorithm RSA -out api_private_key -pkeyopt rsa_keygen_bits:2048

# Generate required public key for api service
api_public_key:
	@$(KEYGEN) rsa -in api_private_key -out api_public_key -pubout

# Generate required private key for ssh service
ssh_private_key:
	@$(KEYGEN) genpkey -algorithm RSA -out ssh_private_key -pkeyopt rsa_keygen_bits:2048

.PHONY: keygen
# Generate required keys
keygen: api_private_key api_public_key ssh_private_key

.PHONY: start
## Start services
start: keygen
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

restart: stop start

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

.PHONY: help
help:
	@cat <<-EOF
		Available commands:
		
		build      Build all services (append "SERVICE=<service>" to build a specific one)
		start      Start services
		stop       Stop services
		EOF

.DEFAULT_GOAL := help
