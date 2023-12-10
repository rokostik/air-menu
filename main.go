package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/caseymrm/menuet"
	"github.com/rokostik/air-menu/api"
)

var apiClient *api.Client
var devices []*api.Device
var selectedDeviceId string
var sensorData *api.SensorData
var selectedSensor string = menuet.Defaults().String("selected-sensor")
var currentError error

var stringFncs = map[string]func(*api.SensorData) string{
	"co2": func(data *api.SensorData) string {
		return fmt.Sprintf("CO2: %d ppm", int(data.Co2))
	},
	"pm25": func(data *api.SensorData) string {
		return fmt.Sprintf("PM2.5: %d µg/m3", int(data.Pm25))
	},
	"temp": func(data *api.SensorData) string {
		return fmt.Sprintf("Temp.: %.1f °C", data.Temp)
	},
	"humidity": func(data *api.SensorData) string {
		return fmt.Sprintf("Hum.: %.1f%%", data.Humidity)
	},
	"radon": func(data *api.SensorData) string {
		return fmt.Sprintf("Radon: %.1f Bq/m3", data.RadonShortTermAvg)
	},
	"pm1": func(data *api.SensorData) string {
		return fmt.Sprintf("PM1: %d µg/m3", int(data.Pm1))
	},
	"voc": func(data *api.SensorData) string {
		return fmt.Sprintf("VOC: %d ppb", int(data.Voc))
	},
}

func setFocusedSensor(sensor string) {
	selectedSensor = sensor
	menuet.Defaults().SetString("selected-sensor", sensor)
	setMenuState()
}

func setMenuState() {
	title := ""
	if sensorData != nil && selectedSensor != "" {
		stringFnc, ok := stringFncs[selectedSensor]
		if ok {
			title += stringFnc(sensorData)
		}
	}
	if currentError != nil {
		title += "(err)"
	}
	menuet.App().SetMenuState(&menuet.MenuState{
		Image: "icon.png",
		Title: title,
	})
	menuet.App().MenuChanged()
}

func refreshData(reccuring bool) {
	for {
		if apiClient == nil {
			time.Sleep(1 * time.Second)
			continue
		}
		if selectedDeviceId == "" {
			var err error
			devices, err = apiClient.GetDevices()
			if err != nil {
				currentError = fmt.Errorf("error getting devices: %s", err)
				setMenuState()
				time.Sleep(1 * time.Minute)
				continue
			}
			selectedDeviceId = devices[0].Id
		}
		var err error
		sensorData, err = apiClient.GetSensorData(selectedDeviceId)
		if err != nil {
			currentError = fmt.Errorf("error getting data: %s", err)
			setMenuState()
			time.Sleep(1 * time.Minute)
			continue
		}
		currentError = nil
		setMenuState()
		if !reccuring {
			return
		}
		time.Sleep(5 * time.Minute)
	}
}

func sensorMenuItem(sensor string) menuet.MenuItem {
	return menuet.MenuItem{
		Text: stringFncs[sensor](sensorData),
		Clicked: func() {
			if selectedSensor == sensor {
				setFocusedSensor("")
			} else {
				setFocusedSensor(sensor)
			}
		},
		State: selectedSensor == sensor,
	}
}

func menuItems() []menuet.MenuItem {
	var menuData []menuet.MenuItem
	if sensorData != nil {
		menuData = []menuet.MenuItem{
			sensorMenuItem("co2"),
			sensorMenuItem("pm25"),
			sensorMenuItem("temp"),
			sensorMenuItem("humidity"),
			sensorMenuItem("radon"),
			sensorMenuItem("pm1"),
			sensorMenuItem("voc"),
			{
				Text: fmt.Sprintf("Updated %s", timeAgo(sensorData.Time)),
			},
		}
	} else {
		menuData = []menuet.MenuItem{
			{Text: "Getting data..."},
		}
	}

	var menuDevices []menuet.MenuItem
	if len(devices) > 0 {
		var menuDevicesChildren []menuet.MenuItem
		for _, device := range devices {
			menuDevicesChildren = append(menuDevices, menuet.MenuItem{
				Text: device.ProductName,
				Clicked: func() {
					selectedDeviceId = device.Id
				},
				State: selectedDeviceId == device.Id,
			})
		}
		menuDevices = []menuet.MenuItem{
			{
				Text:     "Devices",
				Children: func() []menuet.MenuItem { return menuDevicesChildren },
			},
		}
	} else {
		menuDevices = []menuet.MenuItem{
			{
				Text: "Getting devices...",
			},
		}
	}

	var menuItems []menuet.MenuItem
	menuItems = append(menuItems, menuData...)
	menuItems = append(menuItems, menuet.MenuItem{
		Type: menuet.Separator,
	})
	menuItems = append(menuItems, menuDevices...)
	menuItems = append(menuItems, menuet.MenuItem{
		Type: menuet.Separator,
	})
	if currentError != nil {
		menuItems = append(menuItems, menuet.MenuItem{
			Text: currentError.Error(),
		})
	}
	menuItems = append(menuItems, menuet.MenuItem{
		Text: "Reset credentials",
		Clicked: func() {
			createClient(true)
			refreshData(false)
		},
	})

	return menuItems
}

func createClient(forceNew bool) {
	clientId := menuet.Defaults().String("client-id")
	clientSecret := menuet.Defaults().String("client-secret")

	if forceNew || (clientId == "" || clientSecret == "") {
		for {
			res := menuet.App().Alert(menuet.Alert{
				MessageText: "Please input the Airthings API client ID and secret",
				Inputs:      []string{"Client ID", "Client secret"},
				Buttons:     []string{"Ok", "Cancel"},
			})
			if res.Button == 0 {
				clientId = strings.TrimSpace(res.Inputs[0])
				clientSecret = strings.TrimSpace(res.Inputs[1])
				if clientId == "" || clientSecret == "" {
					continue
				} else {
					break
				}
			} else {
				return
			}
		}
	}

	menuet.Defaults().SetString("client-id", clientId)
	menuet.Defaults().SetString("client-secret", clientSecret)
	devices = nil
	sensorData = nil
	apiClient = api.NewClient(clientId, clientSecret)
}

func timeAgo(unixTimestamp int) string {
	t := time.Unix(int64(unixTimestamp), 0)
	duration := time.Since(t)

	minutes := int(duration.Minutes())
	if minutes < 60 {
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	}
	hours := minutes / 60
	if hours < 24 {
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	}
	days := hours / 24
	if days == 1 {
		return "1 day ago"
	}
	return fmt.Sprintf("%d days ago", days)
}

func main() {
	go createClient(false)
	go refreshData(true)
	setMenuState()
	app := menuet.App()
	app.Name = "AirMenu"
	app.Label = "com.github.rokostik.air-menu"
	app.Children = menuItems

	app.RunApplication()
}
