build:
	go build -o gitops


install: build
	sudo cp gitops /usr/local/bin/;
