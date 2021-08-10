build:
	go build -o gops


install: build
	sudo cp gops /usr/local/bin/;
