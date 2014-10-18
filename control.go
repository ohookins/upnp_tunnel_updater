package main

import (
	"bytes"
	"encoding/xml"
	"io/ioutil"
	"log"
	"net/http"
)

func controlMessage() *bytes.Buffer {
	msg := &bytes.Buffer{}
	msg.WriteString("<?xml version=\"1.0\" encoding=\"utf-8\"?>")
	msg.WriteString("<s:Envelope")
	msg.WriteString("  xmlns:s=\"http://schemas.xmlsoap.org/soap/envelope/\"")
	msg.WriteString("  s:encodingStyle=\"http://schemas.xmlsoap.org/soap/encoding/\">")
	msg.WriteString("<s:Body>")
	msg.WriteString("  <u:GetExternalIPAddress xmlns:u=\"urn:schemas-upnp-org:service:WANIPConnection:1\" />")
	msg.WriteString("</s:Body>")
	msg.WriteString("</s:Envelope>")
	return msg
}

type AddressResponse struct {
	XMLName              xml.Name `xml:"GetExternalIPAddressResponse"`
	NewExternalIPAddress string
}
type ResponseBody struct {
	XMLName         xml.Name        `xml:"Body"`
	AddressResponse AddressResponse `xml:"GetExternalIPAddressResponse"`
}
type ResponseEnvelope struct {
	XMLName      xml.Name     `xml:"Envelope"`
	ResponseBody ResponseBody `xml:"Body"`
}

func retrieveWANIP(controlPoint string) string {
	msg := controlMessage()
	req, err := http.NewRequest("POST", controlPoint, msg)
	if err != nil {
		log.Printf("error creating control request: %s", err)
		return ""
	}

	// Required control message headers for SOAP request
	req.Header.Add("Content-Length", string(msg.Len()))
	req.Header.Add("Soapaction", "urn:schemas-upnp-org:service:WANIPConnection:1#GetExternalIPAddress")
	req.Header.Add("Content-Type", "text/xml; charset=\"utf-8\"")

	log.Printf("sending control request to %s", controlPoint)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("error posting control message: %s", err)
	}
	log.Printf("received GetExternalIPAddress response with status code %d", res.StatusCode)
	if res.StatusCode != 200 {
		return ""
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Printf("error reading response body: %s", err)
	}

	responseEnvelope := ResponseEnvelope{}
	err = xml.Unmarshal(body, &responseEnvelope)
	if err != nil {
		log.Printf("error unmarshaling response body: %s", err)
	}

	return responseEnvelope.ResponseBody.AddressResponse.NewExternalIPAddress
}
