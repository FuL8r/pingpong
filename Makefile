# Local developer convenience only, this is not a CI/CD pipeline
.PHONY: help build test race vet cover run clean image push tf-init tf-ensure-ecr deploy destroy

BINARY := bin/pingpong
PKG    := ./...

help: 
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN{FS=":.*?## "}{printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'

build: 
	go build -o $(BINARY) ./cmd/pingpong

test: 
	go test $(PKG)

race: 
	go test -race $(PKG)

vet: 
	go vet $(PKG)

cover: 
	go test -cover $(PKG)

run: build 
	$(BINARY)

clean: 
	rm -rf bin

#AWS deploy
REGION   ?= eu-central-1
TAG      ?= $(shell git rev-parse --short HEAD)
TF_DIR   := deploy/terraform
ACCOUNT   = $(shell aws sts get-caller-identity --query Account --output text)
ECR_URL   = $(ACCOUNT).dkr.ecr.$(REGION).amazonaws.com/pingpong

image: 
	docker build -t pingpong:$(TAG) .

push: image 
	aws ecr get-login-password --region $(REGION) | docker login --username AWS --password-stdin $(ECR_URL)
	docker tag pingpong:$(TAG) $(ECR_URL):$(TAG)
	docker push $(ECR_URL):$(TAG)

tf-init:
	cd $(TF_DIR) && terraform init \
		-backend-config="bucket=$(STATE_BUCKET)" \
		-backend-config="key=pingpong/app-runner.tfstate" \
		-backend-config="region=$(REGION)" \
		-backend-config="dynamodb_table=$(LOCK_TABLE)"

tf-ensure-ecr: 
	cd $(TF_DIR) && terraform apply -auto-approve -target=aws_ecr_repository.pingpong -var image_tag=$(TAG) -var region=$(REGION)

deploy: 
	$(MAKE) tf-ensure-ecr TAG=$(TAG) REGION=$(REGION)
	$(MAKE) push TAG=$(TAG) REGION=$(REGION)
	cd $(TF_DIR) && terraform apply -auto-approve -var image_tag=$(TAG) -var region=$(REGION) $(TF_ARGS)

destroy: 
	cd $(TF_DIR) && terraform destroy -auto-approve -var image_tag=$(TAG) -var region=$(REGION) $(TF_ARGS)
