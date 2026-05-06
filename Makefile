.PHONY: build clean release

BINARY_NAME="air-menu"
FULL_PATH="AirMenu.app/Contents/MacOS/$(BINARY_NAME)"
RELEASE_ZIP="AirMenu.zip"

build:
	go build -o $(FULL_PATH)

release: build
	codesign --force --deep --sign - AirMenu.app
	codesign --verify --deep --strict --verbose=4 AirMenu.app
	ditto -c -k --keepParent AirMenu.app $(RELEASE_ZIP)

clean:
	if [ -f $(FULL_PATH) ] ; then rm $(FULL_PATH) ; fi
	if [ -f $(RELEASE_ZIP) ] ; then rm $(RELEASE_ZIP) ; fi
	if [ -d AirMenu.app/Contents/_CodeSignature ] ; then rm -r AirMenu.app/Contents/_CodeSignature ; fi
