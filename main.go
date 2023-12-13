package main

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/caseymrm/menuet"
	"github.com/rokostik/air-menu/api"
)

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

type AppState struct {
	mu               sync.RWMutex
	apiClient        *api.Client
	devices          []*api.Device
	selectedDeviceId string
	sensorData       *api.SensorData
	selectedSensor   string
	currentError     error
}

func setFocusedSensor(appState *AppState, sensor string) {
	appState.selectedSensor = sensor
	menuet.Defaults().SetString("selected-sensor", sensor)
	setMenuState(appState)
}

func setMenuState(appState *AppState) {
	title := ""
	if appState.sensorData != nil && appState.selectedSensor != "" {
		stringFnc, ok := stringFncs[appState.selectedSensor]
		if ok {
			title += stringFnc(appState.sensorData)
		}
	}
	if appState.currentError != nil {
		title += "(err)"
	}
	menuet.App().SetMenuState(&menuet.MenuState{
		Image: "icon.png",
		Title: title,
	})
	menuet.App().MenuChanged()
}

func refreshData(appState *AppState, reccuring bool) {
	for {
		appState.mu.Lock()

		if appState.apiClient == nil {
			appState.mu.Unlock()
			time.Sleep(1 * time.Second)
			continue
		}
		if appState.selectedDeviceId == "" {
			devices, err := appState.apiClient.GetDevices()
			if err != nil {
				appState.currentError = fmt.Errorf("error getting devices: %s", err)
				setMenuState(appState)
				appState.mu.Unlock()
				time.Sleep(1 * time.Minute)
				continue
			}
			appState.devices = devices
			appState.selectedDeviceId = devices[0].Id
		}
		sensorData, err := appState.apiClient.GetSensorData(appState.selectedDeviceId)
		if err != nil {
			appState.currentError = fmt.Errorf("error getting data: %s", err)
			setMenuState(appState)
			appState.mu.Unlock()
			time.Sleep(1 * time.Minute)
			continue
		}
		appState.sensorData = sensorData
		appState.currentError = nil
		setMenuState(appState)
		appState.mu.Unlock()
		if !reccuring {
			return
		}
		time.Sleep(5 * time.Minute)
	}
}

func sensorMenuItem(appState *AppState, sensor string) menuet.MenuItem {
	return menuet.MenuItem{
		Text: stringFncs[sensor](appState.sensorData),
		Clicked: func() {
			if appState.selectedSensor == sensor {
				setFocusedSensor(appState, "")
			} else {
				setFocusedSensor(appState, sensor)
			}
		},
		State: appState.selectedSensor == sensor,
	}
}

func menuItemsFunc(appState *AppState) func() []menuet.MenuItem {
	return func() []menuet.MenuItem {
		appState.mu.RLock()
		defer appState.mu.RUnlock()

		var menuData []menuet.MenuItem
		if appState.sensorData != nil {
			menuData = []menuet.MenuItem{
				sensorMenuItem(appState, "co2"),
				sensorMenuItem(appState, "pm25"),
				sensorMenuItem(appState, "temp"),
				sensorMenuItem(appState, "humidity"),
				sensorMenuItem(appState, "radon"),
				sensorMenuItem(appState, "pm1"),
				sensorMenuItem(appState, "voc"),
				{
					Text: fmt.Sprintf("Updated %s", timeAgo(appState.sensorData.Time)),
				},
			}
		} else {
			menuData = []menuet.MenuItem{
				{Text: "Getting data..."},
			}
		}

		var menuDevices []menuet.MenuItem
		if len(appState.devices) > 0 {
			var menuDevicesChildren []menuet.MenuItem
			for _, device := range appState.devices {
				menuDevicesChildren = append(menuDevices, menuet.MenuItem{
					Text: device.ProductName,
					Clicked: func() {
						appState.selectedDeviceId = device.Id
					},
					State: appState.selectedDeviceId == device.Id,
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
		if appState.currentError != nil {
			menuItems = append(menuItems, menuet.MenuItem{
				Text: appState.currentError.Error(),
			})
		}
		menuItems = append(menuItems, menuet.MenuItem{
			Text: "Reset credentials",
			Clicked: func() {
				createClient(appState, true)
				refreshData(appState, false)
			},
		})

		return menuItems
	}
}

func createClient(appState *AppState, forceNew bool) {
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

	appState.mu.Lock()
	defer appState.mu.Unlock()

	menuet.Defaults().SetString("client-id", clientId)
	menuet.Defaults().SetString("client-secret", clientSecret)
	appState.devices = nil
	appState.sensorData = nil
	appState.apiClient = api.NewClient(clientId, clientSecret)
}

func main() {
	appState := &AppState{}
	go createClient(appState, false)
	go refreshData(appState, true)

	appState.mu.RLock()
	setMenuState(appState)
	appState.mu.RUnlock()

	app := menuet.App()
	app.Name = "AirMenu"
	app.Label = "com.github.rokostik.air-menu"
	app.Children = menuItemsFunc(appState)

	app.RunApplication()
}
