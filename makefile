all:   
	mkdir -p ./target/crt
	openssl genrsa -des3 -out server.key 2048
	openssl req -new -key server.key -out server.csr
	openssl rsa -in server.key -out server.key
	openssl x509 -req -days 365 -in server.csr -signkey server.key -out server.crt
	mv -f server.crt ./target/crt/server.crt
	mv -f server.key ./target/crt/server.key
	cp -f _config.yaml ./target/config.yaml
	rm -rf server.crt
	rm -rf server.key
	rm -rf server.csr
	go build -o target/ecsEquilizer .
tidy:
	go mod tidy

clean:  
	rm -rf ./target
	rm -rf server.crt
	rm -rf server.key