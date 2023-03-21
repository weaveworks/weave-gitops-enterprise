CALENDAR_VERSION=$(shell date +"%Y-%m")

.PHONY: node_modules
node_modules: package.json yarn.lock
	yarn config set network-timeout 300000 && yarn install --prod --frozen-lockfile

build: node_modules $(shell find src -type f)
	REACT_APP_DISABLE_PROGRESSIVE_DELIVERY="$(REACT_APP_DISABLE_PROGRESSIVE_DELIVERY)" REACT_APP_VERSION="$(CALENDAR_VERSION) $(VERSION)" yarn build

# This job assumes that the weave-gitops repo located next to this repo in the filesystem
core-ui:
	cd ../../weave-gitops && \
	npm run build:lib && \
	npm run typedefs && \
	cd ../weave-gitops-enterprise/ui-cra

core-lib:
	rm -rf node_modules/@weaveworks/weave-gitops/
	yarn add ../../weave-gitops/dist
