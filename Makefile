SRC_LIST := src/main.go src/utils.go
CC := go build

gtranslator: $(SRC_LIST)
	$(CC) -o $@ $^

clean:
	$(RM) gtranslator *.exe

install: gtranslator
	sudo cp -f gtranslator /usr/local/bin/gtranslator
	sudo cp -f desktop/gtranslator.desktop /usr/local/share/applications/gtranslator.desktop
	sudo cp -f desktop/gtranslator.svg /usr/share/icons/hicolor/scalable/apps/gtranslator.svg

uninstall: clean
	sudo $(RM) /usr/local/bin/gtranslator
	sudo $(RM) /usr/local/share/applications/gtranslator.desktop
	sudo $(RM) /usr/share/icons/hicolor/scalable/apps/gtranslator.svg
