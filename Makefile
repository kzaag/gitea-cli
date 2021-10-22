build:
	mkdir -p bin
	go build -o bin/gitea

install: build
	sudo cp bin/gitea /usr/local/bin/;
