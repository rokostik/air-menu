# AirMenu

## About

AirMenu is a MacOS menu bar application for displaying data from your [Airthings](https://www.airthings.com/en/) device.

![demo](https://github.com/rokostik/air-menu/blob/master/demo/demo.gif?raw=true)

## Prerequisites

You need to have an Airthings account with an API client. You can create one [here](https://dashboard.airthings.com/integrations/api-integration).

When the app first starts it will ask you for your API client ID and secret. You can change these later in the dropdown menu.

## Instalation

Download the zip file from the [releases](https://github.com/rokostik/air-menu/releases), extract it and copy the AirMenu.app to your Applications folder.

## Builing

To build from source, you need to have [go](https://go.dev/) installed. Then run:

```bash
make build
```

and copy the AirMenu.app to your Applications folder.
