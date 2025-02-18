publish:
	docker build -t $(REGISTRY)/csrvbot/csrvbot:$(TAG) . --push

publish-arm:
	docker buildx build --platform linux/amd64 -t $(REGISTRY)/csrvbot/csrvbot:$(TAG) . --push
