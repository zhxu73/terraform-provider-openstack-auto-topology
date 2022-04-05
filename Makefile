VERSION=0.1
ARCH=linux_amd64
PROVIDER_NAME=openstack-auto-topology
EXECUTABLE_FILENAME=terraform-provider-$(PROVIDER_NAME)
PROVIDER_DIR=$(HOME)/.terraform.d/plugins/terraform.cyverse.org/cyverse/$(PROVIDER_NAME)/$(VERSION)/$(ARCH)

# install the provider executable to home directory
.PHONY: install
install: build
	mkdir -p $(PROVIDER_DIR)
	cp $(EXECUTABLE_FILENAME) $(PROVIDER_DIR)/$(EXECUTABLE_FILENAME)

.PHONY: build
build:
	go build -o $(EXECUTABLE_FILENAME)
