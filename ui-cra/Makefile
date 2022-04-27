CALENDAR_VERSION=$(shell date +"%Y-%m")

node_modules:
	yarn config set network-timeout 300000 && yarn install --prod --frozen-lockfile

build: node_modules $(shell find src -type f)
	REACT_APP_VERSION="$(CALENDAR_VERSION) $(VERSION)" yarn build
