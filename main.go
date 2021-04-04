package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

var (
	RPIFanStatusEndpoint  string
	RPITurnOnFanEndpoint  string
	RPITurnOffFanEndpoint string
)
var myClient = &http.Client{Timeout: 10 * time.Second}

type RPITepmStatus struct {
	FanIsWorking   string  `json:"current_fan_state"`
	RPICurrentTemp float64 `json:"current_temp"`
}
type RPIResponseStatus struct {
	ResponseMsg string `json:"status"`
}

func main() {
	RPIFanControllerAddress := os.Getenv("RPI_FAN_CONTROLLER_ADDRESS")
	u, _ := url.Parse(RPIFanControllerAddress)
	RPIUrl := u.String()
	if RPIUrl == "" {
		log.Panicln("pass environment variable RPI_FAN_CONTROLLER_ADDRESS")
	}

	RPIFanStatusEndpoint = fmt.Sprintf("%s/%s", RPIUrl, "")
	RPITurnOnFanEndpoint = fmt.Sprintf("%s/%s", RPIUrl, "on")
	RPITurnOffFanEndpoint = fmt.Sprintf("%s/%s", RPIUrl, "off")

	checkRpiTemperature()
}

func checkRpiTemperature() {
	resp := RPITepmStatus{}
	for {
		status := ""
		err := getRPIStatus(&resp)
		if err != nil {
			log.Printf("ERROR - during fetching status - %s", err)
			time.Sleep(time.Second * 10)
			continue
		}
		currentRPITemp := resp.RPICurrentTemp
		if resp.FanIsWorking == "on" {
			status, err = fanIsOnLogic(currentRPITemp)
		} else {
			status, err = fanIsOffLogic(currentRPITemp)
		}
		if err != nil {
			log.Printf("ERROR - during fan controlling - %s", err)
			time.Sleep(time.Second * 10)
			continue
		}
		if status == "ok" {
			log.Println("INFO - RPI fan triggering succeeded")
		} else if status == "error" {
			log.Println("INFO - RPI fan triggering failed")
		}
		time.Sleep(time.Minute)
	}
}

func getRPIStatus(rpiTempStatus *RPITepmStatus) error {
	r, err := myClient.Get(RPIFanStatusEndpoint)
	if err != nil {
		return err
	}
	defer func() {
		err := r.Body.Close()
		if err != nil {
			log.Printf("ERROR - during closing connection - %s", err)
		}
	}()
	return json.NewDecoder(r.Body).Decode(rpiTempStatus)
}

func fanIsOnLogic(currentRPITemp float64) (string, error) {
	var err error
	if currentRPITemp < 50.0 {
		rpiResponseStatus := RPIResponseStatus{}
		err = triggerFan(&rpiResponseStatus, RPITurnOffFanEndpoint)
		log.Println("INFO - triggered RPI fan to turn off")
		status := rpiResponseStatus.ResponseMsg
		return status, err
	}
	log.Println("INFO - RPI temperature is under control")
	log.Println("INFO - RPI fan is turn on")
	return "", nil
}

func fanIsOffLogic(currentRPITemp float64) (string, error) {
	var err error
	if currentRPITemp > 60.0 {
		rpiResponseStatus := RPIResponseStatus{}
		err = triggerFan(&rpiResponseStatus, RPITurnOnFanEndpoint)
		if err != nil {
			return "", err
		}
		log.Println("INFO - triggered RPI fan to turn on")
		status := rpiResponseStatus.ResponseMsg
		return status, err
	}
	log.Println("INFO - RPI temperature is under control")
	log.Println("INFO - RPI fan is turn off")
	return "", nil
}

func triggerFan(rpiResponseStatus *RPIResponseStatus, endpoint string) error {
	r, err := myClient.Get(endpoint)
	if err != nil {
		return err
	}
	defer func() {
		err := r.Body.Close()
		if err != nil {
			log.Printf("ERROR - during closing connection - %s", err)
		}
	}()
	err = json.NewDecoder(r.Body).Decode(rpiResponseStatus)
	return err
}
