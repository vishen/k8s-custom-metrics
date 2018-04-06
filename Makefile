
deps:
	glide install -v

docker_build:
	docker build -t k8s-custom-metrics .
