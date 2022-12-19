CALENDAR_VERSION=$(shell date +"%Y-%m")
TIER=enterprise
TEST1=test1

.PHONY: node_modules
node_modules: package.json yarn.lock
	yarn config set network-timeout 300000 && yarn install --prod --frozen-lockfile

build: node_modules $(shell find src -type f)
	export TIER=$(TIER) && export TEST1=$(TEST1) && REACT_APP_DISABLE_PROGRESSIVE_DELIVERY="$(REACT_APP_DISABLE_PROGRESSIVE_DELIVERY)" REACT_APP_VERSION="$(CALENDAR_VERSION) $(VERSION)" yarn build

core-lib:
	rm -rf node_modules/@weaveworks/weave-gitops/
	yarn add ../../weave-gitops/dist
