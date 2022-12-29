all: build test
default: build

docker_build_api:
	docker build  -t codeowners_manager_api . -f ./Dockerfile_api

docker_serve_api:
	docker run --name codeowners_manager_api -p 8080:8080 \
		--env aws_region=$$aws_region \
		--env AWS_ACCESS_KEY_ID=$$AWS_ACCESS_KEY_ID \
		--env AWS_SECRET_ACCESS_KEY=$$AWS_SECRET_ACCESS_KEY \
		--env codeowners_host_table=$$codeowners_host_table \
		--env codeowners_repositoryowner_table=$$codeowners_repositoryowner_table \
		--env codeowners_ttl_minutes=$$codeowners_ttl_minutes \
 		--rm codeowners_manager_api

docker_build_loader:
	docker build  -t codeowners_manager_loader . -f ./Dockerfile_loader

docker_serve_loader:
	docker run --name codeowners_manager_loader \
		--env aws_region=$$aws_region \
		--env AWS_ACCESS_KEY_ID=$$AWS_ACCESS_KEY_ID \
		--env AWS_SECRET_ACCESS_KEY=$$AWS_SECRET_ACCESS_KEY \
		--env codeowners_host_table=$$codeowners_host_table \
		--env codeowners_repositoryowner_table=$$codeowners_repositoryowner_table \
		--env codeowners_ttl_minutes=$$codeowners_ttl_minutes \
 		--rm codeowners_manager_api

build:
	go build ./...

test:
	go test ./...