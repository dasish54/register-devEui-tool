# register-devEui-tool

Usage
go run main.go

Requirements
- a) The application must return every DevEUI that it registers with the LoRaWAN
provider (e.g. if the application is killed it must wait for in-flight requests to finish
otherwise we would have registered those DevEUIs but would not be using them)
- b) It must handle user interrupts gracefully (SIGINT)
- c) It must register exactly 100 DevEUIs (no more) with the provider (to avoid paying
for DevEUIs that we do not use)
- d) It should make multiple requests concurrently (but there must never be more than
10 requests in-flight, to avoid throttling)

Improvements
- Re-think the concurrency pattern, the current retry method doesn't seem like the cleanest solution.
- When the api responds saying the devEui has already been registered, it generates an new devEui to try, the newly generated devEui doesn't get checked with the other 
99 devEui to ensure uniqueness.
