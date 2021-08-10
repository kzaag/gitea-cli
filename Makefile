run:	
	mkdir -p bin
	go build -o bin/gitea_ops *.go
	cd bin; ./gitea_ops
