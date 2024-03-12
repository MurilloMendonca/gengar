all:
	go build -o gengar
clean:
	rm -f gengar
install:
	cp gengar /usr/local/bin
