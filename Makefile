GOCMD=$(shell which go)
GOBUILD=$(GOCMD) build
BUILDPATH=$(CURDIR)
BINARYPATH=$(BUILDPATH)/bin
BINARYNAME=arrowcloudapi
SOURCECODE=.
DOCKERCMD=$(shell which docker)
DOCKERBUILD=$(DOCKERCMD) build
DOCKERPUSH=$(DOCKERCMD) push
DOCKERIMAGENAME=services-registry.cloudapp-enterprise-preprod.appctest.com:5000/ypjin/arrowcloudapi
VERSIONTAG=test

compile:
	@echo "compiling binary for arrowcloudapi..."
	@echo GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARYPATH)/$(BINARYNAME) $(SOURCECODE)
	@env GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARYPATH)/$(BINARYNAME) $(SOURCECODE)
	@echo "Done."

image:
	@echo "building docker image for alpine..."
	$(DOCKERBUILD) -f $(BUILDPATH)/Dockerfile -t $(DOCKERIMAGENAME):$(VERSIONTAG) .
	@echo "Done."

push:
	@echo "pushing docker image for alpine..."
	$(DOCKERPUSH) $(DOCKERIMAGENAME):$(VERSIONTAG)
	@echo "Done."