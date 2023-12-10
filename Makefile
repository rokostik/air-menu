.PHONY: build clean

BINARY_NAME="air-menu"
FULL_PATH="AirMenu.app/Contents/MacOS/$(BINARY_NAME)"

build:
	go build -o $(FULL_PATH)

clean:
	if [ -f $(FULL_PATH) ] ; then rm $(FULL_PATH) ; fi
