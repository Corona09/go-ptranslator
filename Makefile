SRC_LIST := src/main.go src/utils.go
CC := go build

gtranslator: $(SRC_LIST)
	$(CC) -o $@ $^

clean:
	$(RM) gtranslator *.exe

install: gtranslator
	cp gtranslator /usr/lcoal/bin/gtranslator

uninstall: clean
	$(RM) /usr/local/bin/gtranslator
