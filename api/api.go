package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"golang.org/x/oauth2/clientcredentials"
)

const tokenUrl = "https://accounts-api.airthings.com/v1/token"
const baseUrl = "https://ext-api.airthings.com/v1"

type Device struct {
	Id          string   `json:"id"`
	DeviceType  string   `json:"deviceType"`
	Sensors     []string `json:"sensors"`
	ProductName string   `json:"productName"`
}

type SensorData struct {
	Time              int     `json:"time"`
	Battery           int     `json:"battery"`
	Co2               float64 `json:"co2"`
	Humidity          float64 `json:"humidity"`
	Pm1               float64 `json:"pm1"`
	Pm25              float64 `json:"pm25"`
	Pressure          float64 `json:"pressure"`
	RadonShortTermAvg float64 `json:"radonShortTermAvg"`
	RelayDeviceType   string  `json:"relayDeviceType"`
	Rssi              int     `json:"rssi"`
	Temp              float64 `json:"temp"`
	Voc               float64 `json:"voc"`
}

type Client struct {
	clientId     string
	clientSecret string
	httpClient   *http.Client
}

func NewClient(clientId string, clientSecret string) *Client {
	conf := clientcredentials.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		TokenURL:     tokenUrl,
	}

	return &Client{
		clientId:     clientId,
		clientSecret: clientSecret,
		httpClient:   conf.Client(context.Background()),
	}
}

func (c Client) GetDevices() ([]*Device, error) {
	res, err := c.httpClient.Get(fmt.Sprintf("%s/devices", baseUrl))
	if err != nil {
		return nil, fmt.Errorf("%s", err.(*url.Error).Err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET devices %d", res.StatusCode)
	}

	var devices struct {
		Devices []*Device `json:"devices"`
	}
	err = json.NewDecoder(res.Body).Decode(&devices)
	if err != nil {
		return nil, err
	}
	return devices.Devices, nil
}

func (c Client) GetSensorData(deviceId string) (*SensorData, error) {
	res, err := c.httpClient.Get(fmt.Sprintf("%s/devices/%s/latest-samples", baseUrl, deviceId))
	if err != nil {
		return nil, fmt.Errorf("%s", err.(*url.Error).Err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET latest-samples %d", res.StatusCode)
	}

	var sensorData struct {
		Data *SensorData `json:"data"`
	}
	err = json.NewDecoder(res.Body).Decode(&sensorData)
	if err != nil {
		return nil, err
	}
	return sensorData.Data, nil
}
