package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type createDevEuiRequest struct {
	DevEui string `json:"devEui"`
}

var src = rand.New(rand.NewSource(time.Now().UnixNano()))
var devEuiAlreadyRegistered = errors.New("devEui has already been registered")

func main() {

	devEuis := generateNumUniqueDevEuis([]string{}, 100)

	var semaphore = make(chan int, 10)

	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, syscall.SIGINT)

	var client = &http.Client{}
	var mutex = &sync.Mutex{}
	var wg sync.WaitGroup

	devEuisRegistered := make([]string, 0, 100)

	for i := 0; i < len(devEuis); i++ {

		select {
		case <-termChan:
			fmt.Println("process terminating: waiting for in-flight requests to complete")
			wg.Wait()
			displayDevEuis(devEuisRegistered)
			return

		default:
		}

		semaphore <- 1
		wg.Add(1)

		go func(devEui string) {
			defer wg.Done()

			maxRetries := 5
			retries := 0
			for {
				err := remoteAddNewDevEui(client, devEui)
				if err == devEuiAlreadyRegistered {
					retries++
					// TODO: this is not going to generate one with a unique last few chars
					devEui = generateDevEui()
				}

				if err == nil {
					mutex.Lock()
					devEuisRegistered = append(devEuisRegistered, devEui)
					mutex.Unlock()
					break
				}

				if retries == maxRetries {
					log.Fatalf("ERROR job")
				}
			}

			<-semaphore
		}(devEuis[i])
	}
	wg.Wait()

	displayDevEuis(devEuisRegistered)

}

func remoteAddNewDevEui(client *http.Client, devEui string) error {
	createDevEuiRequestBody, _ := json.Marshal(createDevEuiRequest{DevEui: devEui})
	resp, err := http.Post("http://europe-west1-machinemax-dev-d524.cloudfunctions.net/sensor-onboarding-sample", "application/json", bytes.NewBuffer(createDevEuiRequestBody))
	if err != nil {
		return err
	}

	if resp.StatusCode == 200 {
		return nil
	} else if resp.StatusCode == 422 {
		return devEuiAlreadyRegistered
	}
	return errors.New("failed to register devEui")
}

func displayDevEuis(devEuis []string) {
	fmt.Printf("number of DevEuis registered: %d", len(devEuis))
	for i, v := range devEuis {
		fmt.Printf("\n %d : %s", i, v)
	}
}

func generateNumUniqueDevEuis(currentDevEuis []string, totalNumberOfDevEuisRequired int) []string {

	devEuisToGenerate := totalNumberOfDevEuisRequired - len(currentDevEuis)
	// generate new devEuis
	for i := 0; i < devEuisToGenerate; i++ {
		currentDevEuis = append(currentDevEuis, generateDevEui())
	}

	uniqueDevEuis := make(map[string]string, totalNumberOfDevEuisRequired)

	for _, devEui := range currentDevEuis {
		if _, exists := uniqueDevEuis[devEui[11:]]; exists {
			// no-op as last 5 chars clash with pre-existing devEui
		} else {
			uniqueDevEuis[devEui[11:]] = devEui
		}
	}

	uniqueDevEuisSlice := make([]string, 0, len(uniqueDevEuis))

	for _, value := range uniqueDevEuis {
		uniqueDevEuisSlice = append(uniqueDevEuisSlice, value)
	}

	if len(uniqueDevEuisSlice) == totalNumberOfDevEuisRequired {
		return uniqueDevEuisSlice
	}

	return generateNumUniqueDevEuis(uniqueDevEuisSlice, totalNumberOfDevEuisRequired)

}

func generateDevEui() string {
	b := make([]byte, (16 / 2))

	if _, err := src.Read(b); err != nil {
		panic(err)
	}

	return hex.EncodeToString(b)

}
