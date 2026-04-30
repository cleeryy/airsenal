BIN     := bin/airsenal
CMD     := ./cmd/airsenal
IMAGE   := airsenal:latest

.PHONY: build run test lint docker-build clean update-cheats

build:
	go build -o $(BIN) $(CMD)

run:
	go run $(CMD)

test:
	go test ./...

lint:
	go vet ./...

docker-build:
	docker build -t $(IMAGE) .

clean:
	rm -rf bin/

update-cheats:
	git submodule update --remote --merge
