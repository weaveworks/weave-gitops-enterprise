CALENDAR_VERSION=$(shell date +"%Y-%m")

node_modules: package.json yarn.lock
	yarn config set network-timeout 300000 && yarn install --prod --frozen-lockfile

build: node_modules $(shell find src -type f)
	REACT_APP_VERSION="$(CALENDAR_VERSION) $(VERSION)" yarn build

core-lib:
	rm -rf node_modules/@weaveworks/weave-gitops/
	yarn add ../../weave-gitops/dist
