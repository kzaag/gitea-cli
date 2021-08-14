build:
	go build -o gitea

install: build
	sudo cp gitea /usr/local/bin/;
