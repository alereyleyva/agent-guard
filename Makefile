# Misc
.DEFAULT_GOAL = help
.PHONY        : help ruler

## —— Agent Guard Makefile —————————————————————————————————————————
help: ## Outputs this help screen
	@grep -E '(^[a-zA-Z0-9\./_-]+:.*?##.*$$)|(^##)' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}{printf "\033[32m%-30s\033[0m %s\n", $$1, $$2}' | sed -e 's/\[32m##/[33m/'

ruler: ## Apply ruler configuration
	@bunx @intellectronica/ruler@latest apply --local-only
