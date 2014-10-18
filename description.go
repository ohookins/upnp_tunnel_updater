package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

// Define structs necessary for unmarshalling device description
type Service struct {
	XMLName     xml.Name `xml:"service"`
	ServiceType string   `xml:"serviceType"`
	ServiceId   string   `xml:"serviceId"`
	SCPDURL     string
	ControlURL  string `xml:"controlURL"`
	EventSubURL string `xml:"eventSubURL"`
}
type ServiceList struct {
	XMLName xml.Name `xml:"serviceList"`
	Service Service  `xml:"service"`
}
type SecondLevelDevice struct {
	XMLName     xml.Name    `xml:"device"`
	ServiceList ServiceList `xml:"serviceList"`
}
type SecondLevelDeviceList struct {
	XMLName           xml.Name          `xml:"deviceList"`
	SecondLevelDevice SecondLevelDevice `xml:"device"`
}
type FirstLevelDevice struct {
	XMLName               xml.Name              `xml:"device"`
	SecondLevelDeviceList SecondLevelDeviceList `xml:"deviceList"`
}
type RootDeviceList struct {
	XMLName          xml.Name         `xml:"deviceList"`
	FirstLevelDevice FirstLevelDevice `xml:"device"`
}
type RootDevice struct {
	XMLName        xml.Name       `xml:"device"`
	RootDeviceList RootDeviceList `xml:"deviceList"`
}
type serviceDescription struct {
	XMLName    xml.Name   `xml:"root"`
	RootDevice RootDevice `xml:"device"`
}

func getWANControlPoint(msg map[string]string) (string, error) {
	if _, present := msg["location"]; !present {
		return "", fmt.Errorf("message had no 'location' header")
	}

	// Get base query endpoint from the Location header
	locationURL, err := url.Parse(msg["location"])
	if err != nil {
		return "", fmt.Errorf("location header appears to be invalid: %s", err)
	}
	queryEndpoint := fmt.Sprintf("%s://%s", locationURL.Scheme, locationURL.Host)

	res, err := http.Get(msg["location"])
	if err != nil {
		return "", fmt.Errorf("error retrieving service description: %s", err)
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("error reading service description body: %s", err)
	}

	desc := serviceDescription{}
	err = xml.Unmarshal(body, &desc)
	if err != nil {
		return "", fmt.Errorf("error unmarshaling service description response: %s", err)
	}

	// FIXME: Do something with this monstrosity!
	controlURL := desc.RootDevice.RootDeviceList.FirstLevelDevice.SecondLevelDeviceList.SecondLevelDevice.ServiceList.Service.ControlURL
	return fmt.Sprintf("%s%s", queryEndpoint, controlURL), nil
}
