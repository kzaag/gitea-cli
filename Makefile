build:
	mkdir -p bin
	go build -o bin/gitea

install: build
	sudo cp gitea /usr/local/bin/;
