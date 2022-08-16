package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
)

func main() {
	washers, err := getMachines("LAVADORA")
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}
	jsonWashers, _ := json.Marshal(washers)
	fmt.Println(string(jsonWashers))

	fmt.Println("")

	dryers, err := getMachines("SECADORA")
	if err != nil {
		log.Fatal(err)
	}

	jsonDryers, _ := json.Marshal(dryers)
	fmt.Println(string(jsonDryers))
}

func getMachines(machineType string) (map[string]*Machine, error) {
	err := playwright.Install()
	if err != nil {
		log.Fatalf("Error installing playwright: %v", err)
	}
	// Launch browser
	pw, err := playwright.Run()
	if err != nil {
		errString := fmt.Sprintf("could not start playwright: %v", err)
		err = errors.New(errString)
		return nil, err
	}
	var headless = true
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{Args: []string{"--disable-gpu"}, Headless: &headless})
	if err != nil {
		errString := fmt.Sprintf("could not launch browser: %v", err)
		err = errors.New(errString)
		return nil, err
	}
	defer browser.Close()
	defer pw.Stop()
	page, err := browser.NewPage()
	if err != nil {
		errString := fmt.Sprintf("could not create page: %v", err)
		err = errors.New(errString)
		return nil, err
	}
	// Go to web
	_, err = page.Goto("<WEBSITE URL>", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	if err != nil {
		errString := fmt.Sprintf("could not goto: %v", err)
		err = errors.New(errString)
		return nil, err
	}
	// LOGIN PAGE
	fmt.Println("LOGIN PAGE")
	fmt.Println("Signing in...")
	// Obtain login frame
	el, err := page.WaitForSelector("html > frameset > frame:nth-child(2)")
	if err != nil {
		errString := fmt.Sprintf("frame not found: %v", err)
		err = errors.New(errString)
		return nil, err
	}
	frame, err := el.ContentFrame()
	if err != nil {
		errString := fmt.Sprintf("element is not a frame: %v", err)
		err = errors.New(errString)
		return nil, err
	}
	// Login: user > psswd > accept
	err = frame.Click("#txtusuario")
	if err != nil {
		errString := fmt.Sprintf("could not click: %v", err)
		err = errors.New(errString)
		return nil, err
	}
	err = frame.Type("#txtusuario", "<YOUR USERNAME>")
	if err != nil {
		errString := fmt.Sprintf("could not type: %v", err)
		err = errors.New(errString)
		return nil, err
	}
	err = frame.Click("#txtContraseña")
	if err != nil {
		errString := fmt.Sprintf("could not click: %v", err)
		err = errors.New(errString)
		return nil, err
	}
	err = frame.Type("#txtContraseña", "<YOUR PASSWORD>")
	if err != nil {
		errString := fmt.Sprintf("could not type: %v", err)
		err = errors.New(errString)
		return nil, err
	}
	err = frame.Press("#btnAcpetar", "Enter")
	if err != nil {
		errString := fmt.Sprintf("could not create press: %v", err)
		err = errors.New(errString)
		return nil, err
	}

	// HOME PAGE
	fmt.Println("HOME PAGE")
	fmt.Println("Click on configuration button")
	frame.Click("#lblMenu > ul > li:nth-child(2) > a")

	frame.WaitForTimeout(1000)
	frame.WaitForLoadState()
	// CONFIGURATION
	fmt.Println("CONFIGURATION")
	fmt.Println("Scraping data...")
	values := []string{"25"}
	frame.SelectOption("#example_length > label > span.custom-select > select", playwright.SelectOptionValues{Values: &values})
	names, err := frame.QuerySelectorAll("tbody > tr > td:nth-child(1)")
	if len(names) == 0 {
		errString := "could not get entries"
		err = errors.New(errString)
		return nil, err
	}
	if err != nil {
		errString := fmt.Sprintf("could not get entries: %v", err)
		err = errors.New(errString)
		return nil, err
	}
	ids, err := frame.QuerySelectorAll("tbody > tr > td:nth-child(3)")
	if err != nil {
		errString := fmt.Sprintf("could not get entries: %v", err)
		err = errors.New(errString)
		return nil, err
	}
	prices, err := frame.QuerySelectorAll("tbody > tr > td:nth-child(4)")
	if err != nil {
		errString := fmt.Sprintf("could not get entries: %v", err)
		err = errors.New(errString)
		return nil, err
	}

	machinesMap := make(map[string]*Machine)
	for i := 0; i < len(names); i++ {
		name, err := names[i].TextContent()
		if err != nil {
			errString := fmt.Sprintf("could not get text content: %v", err)
			err = errors.New(errString)
			return nil, err
		}
		if strings.Contains(name, machineType) {
			idString, err := ids[i].TextContent()
			if err != nil {
				errString := fmt.Sprintf("could not get text content: %v", err)
				err = errors.New(errString)
				return nil, err
			}
			priceString, err := prices[i].TextContent()
			if err != nil {
				errString := fmt.Sprintf("could not get text content: %v", err)
				err = errors.New(errString)
				return nil, err
			}

			id, err := strconv.Atoi(idString)
			if err != nil {
				return nil, err
			}
			priceString = strings.Split(priceString, " ")[0]
			priceString = strings.Replace(priceString, ",", ".", 1)
			price, err := strconv.ParseFloat(priceString, 64)
			if err != nil {
				return nil, err
			}

			m := Machine{
				ID:       id,
				Name:     name,
				Status:   "",
				TimeLeft: 0,
				Price:    price,
			}
			machinesMap[name] = &m
		}
	}

	// Go to machines state
	fmt.Println("Click on machines state button")
	frame.Click("#lblMenu > ul > li:nth-child(5) > a")

	frame.WaitForTimeout(1000)
	frame.WaitForLoadState()
	frame.WaitForSelector("#form1")

	// MACHINES STATE
	fmt.Println("MACHINES STATE")
	fmt.Println("Scraping more data...")

	names, err = frame.QuerySelectorAll("div > div.pmd-display2")
	if len(names) == 0 {
		errString := "could not get entries"
		err = errors.New(errString)
		return nil, err
	}
	if err != nil {
		errString := fmt.Sprintf("could not get entries: %v", err)
		err = errors.New(errString)
		return nil, err
	}
	startTimes, err := frame.QuerySelectorAll("div > div.source-semibold.typo-fill-secondary")
	if err != nil {
		errString := fmt.Sprintf("could not get entries: %v", err)
		err = errors.New(errString)
		return nil, err
	}
	statusImgs, err := frame.QuerySelectorAll("div > a > img")
	if err != nil {
		errString := fmt.Sprintf("could not get entries: %v", err)
		err = errors.New(errString)
		return nil, err
	}

	for i := 0; i < len(names); i++ {
		name, err := names[i].TextContent()
		if err != nil {
			errString := fmt.Sprintf("could not get text content: %v", err)
			err = errors.New(errString)
			return nil, err
		}
		if strings.Contains(name, machineType) {

			startTime, err := startTimes[i].TextContent()

			if err != nil {
				errString := fmt.Sprintf("could not get text content: %v", err)
				err = errors.New(errString)
				return nil, err
			}
			var timeLeft int
			if startTime != "-" {
				startTimeSplitted := strings.Split(startTime, ":")
				startTimeHours, err := strconv.Atoi(startTimeSplitted[0])
				if err != nil {
					return nil, err
				}
				startTimeMinutes, err := strconv.Atoi(startTimeSplitted[1])
				if err != nil {
					return nil, err
				}
				startTimeDec := startTimeHours + startTimeMinutes/60.0
				currentTime := time.Now().Format("15:04")
				currentTimeSplitted := strings.Split(currentTime, ":")
				currentTimeHours, err := strconv.ParseFloat(currentTimeSplitted[0], 64)
				if err != nil {
					return nil, err
				}
				currentTimeMinutes, err := strconv.ParseFloat(currentTimeSplitted[1], 64)
				if err != nil {
					return nil, err
				}
				currentTimeDec := currentTimeHours + currentTimeMinutes/60.0
				timeLeft = int((30 - (currentTimeDec - float64(startTimeDec))) * 60)
			} else {
				timeLeft = 0
			}

			statusImg, err := statusImgs[i+1].GetAttribute("src")
			if err != nil {
				errString := fmt.Sprintf("could not get attribute: %v", err)
				err = errors.New(errString)
				return nil, err
			}

			var status string
			switch statusImg {
			case "/imgs/lavlibre.png":
				status = "green"
			case "/imgs/lavocup.png":
				status = "red"
			case "/imgs/lavpend.png":
				status = "blue"
			case "/imgs/seclibre.png":
				status = "green"
			case "/imgs/secocup.png":
				status = "red"
			case "/imgs/secpend.png":
				status = "blue"
			default:
				log.Println("NO CARGA LA IMAGEN DE LA MAQUINA")
			}

			machinesMap[name].TimeLeft = timeLeft
			machinesMap[name].Status = status
		}
	}

	return machinesMap, nil
}
