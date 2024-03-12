all:
	go build -o gengar
clean:
	rm -f gengar
install:
	cp gengar /usr/local/bin
	mkdir -p /usr/local/share/gengar
	cp -r templates /usr/local/share/gengar/templates
uninstall:
	rm -f /usr/local/bin/gengar
	rm -rf /usr/local/share/gengar
