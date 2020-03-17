package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	_ "sort"
	"strconv"

	"github.com/gorilla/mux"
)

const IP_ADDRESS = "http://127.0.0.1:"
const PORT = 12345
const CONTENT_TYPE = "application/json"
const OK = "** 200 OK **"
const ACCEPTED = "** 202 Accepted **"
const NO_CONTENT = "** 204 No content **"
const BAD_REQUEST = "** 400 Bad request **"
const NOT_FOUND = "** 404 Not found **"
const BAD_GATEWAY = "** 502 Bad gateway **"

var groupCount int = 0
var journeyCount int = 0

type Car struct {
	Id      int `json:"id"`
	Seats   int `json:"seats"`
	Engaged bool
}

type CarArray []Car

func (ca CarArray) Len() int {
	return len(ca)
}

func (ca CarArray) Less(i, j int) bool {
	return ca[i].Seats < ca[j].Seats
}

func (ca CarArray) Swap(i, j int) {
	ca[i], ca[j] = ca[j], ca[i]
}

type Group struct {
	Id int `json:"id"`
}

type GroupArray []Group

func (ga GroupArray) Len() int {
	return len(ga)
}

func (ga GroupArray) Less(i, j int) bool {
	return ga[i].Id < ga[j].Id
}

func (ga GroupArray) Swap(i, j int) {
	ga[i], ga[j] = ga[j], ga[i]
}

type Journey struct {
	Id      int `json:"id"`
	Seats   int `json:"seats"`
	GroupId int `json:"gid"`
	CarId   int `json:"cid"`
}

type JourneyArray []Journey

func (ja JourneyArray) Len() int {
	return len(ja)
}

func (ja JourneyArray) Less(i, j int) bool {
	return ja[i].Seats < ja[j].Seats
}

func (ja JourneyArray) Swap(i, j int) {
	ja[i], ja[j] = ja[j], ja[i]
}

type Message struct {
	Body string `json:"message"`
	Info int    `json:"info"`
}

var groups []Group
var cars []Car
var journeys []Journey

// Find a car by ID.
func findCarById(target []Car, targetSize int, cid int) int {
	if targetSize <= 0 || cid < 1 {
		return -1
	}
	/*
		sort.Slice(target, func(i, j int) bool { return target[i].Id <= target[j].Id })
		i := sort.Search(targetSize, func(i int) bool { return target[i].Id >= cid })

		if target[i].Id == cid {
			return i
		}
	*/

	for i, t := range target {
		if t.Id == cid {
			return i
		}
	}

	return -1
}

// Find a car by seats.
func findCarBySeats(target []Car, targetSize int, seats int) int {
	if targetSize <= 0 {
		return -1
	}
	/*
		sort.Sort(CarArray(target))
		maxItem := target[len(target)-1]
		if seats > maxItem.Seats {
			return -1
		}

		cid := sort.Search(targetSize, func(i int) bool { return target[i].Seats >= seats })
		cidOk := cid >= 0 && cid < targetSize
		seatsOk := target[cid].Seats >= seats
		engaged := target[cid].Engaged
		if cidOk && seatsOk && !engaged {
			return target[cid].Id
		}
	*/

	for _, t := range target {
		if t.Seats == seats && !t.Engaged {
			return t.Id
		}
	}

	return -1
}

// Check whether group exists.
func findGroup(target []Group, targetSize int, gid int) bool {
	if targetSize <= 0 {
		return false
	}
	/*
		sort.Sort(GroupArray(target))
		groupId := sort.Search(targetSize, func(i int) bool { return target[i].Id >= gid })

		if groupId < targetSize && target[groupId].Id == gid {
			return true
		}
	*/
	for _, g := range target {
		if g.Id == gid {
			return true
		}
	}

	return false
}

// Find a group by ID.
func findGroupByGid(target []Group, targetSize int, gid int) int {
	if targetSize <= 0 || gid < 1 {
		return -1
	}

	for i, t := range target {
		if t.Id == gid {
			return i
		}
	}

	return -1
}

// Find a journey by ID.
func findJourneyByJid(target []Journey, targetSize int, jid int) int {
	if targetSize <= 0 || jid < 1 {
		return -1
	}
	/*
		sort.Sort(JourneyArray(target))
		journeyId := sort.Search(targetSize, func(i int) bool { return target[i].Id <= jid })
		fmt.Printf("FCBI/journey id: %d\n", journeyId)
		if journeyId < targetSize {
			return journeyId
		}
	*/
	for i, t := range target {
		if t.Id == jid {
			return i
		}
	}

	return -1
}

// Find journey ID by its group ID.
func findJourneyByGroup(target []Journey, targetSize int, gid int) int {
	if targetSize <= 0 {
		return -1
	}
	/*
		sort.Sort(JourneyArray(target))

		journeyId := sort.Search(targetSize, func(i int) bool { return target[i].GroupId <= gid })
		fmt.Printf("FJBG/jid: %d\n", journeyId)
		if journeyId < targetSize && target[journeyId].GroupId == gid {
			return journeyId
		}
	*/

	for _, t := range target {
		if t.GroupId == gid {
			return t.Id
		}
	}

	return -1
}

// Build instance of the Message struct and convert it to JSON.
func messageToJson(msg string, info int) []byte {
	m2s := Message{Body: msg, Info: info}
	s2j, err := json.Marshal(m2s)

	if err != nil {
		panic(err.Error())
	}

	return s2j
}

// Set HTTP status, JSON message, and respond to client.
func respond(w http.ResponseWriter, httpStatus int, contentType string, message []byte) {
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(httpStatus)
	w.Write(message)
}

// It is like a landing page.
func home(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		respond(w, 200, CONTENT_TYPE, messageToJson(OK, -1))
	}
}

// Load the list of available cars in the service and remove all previous data(existing journeys and cars).
func loadCars(w http.ResponseWriter, r *http.Request) {
	if r.Method == "PUT" {
		// Remove existing journeys and cars.
		journeys = journeys[:0]
		cars = cars[:0]

		// Use the curl CLI utility to import data on available cars from local JSON file.
		lines, err := ioutil.ReadAll(r.Body)
		if err != nil {
			respond(w, 502, CONTENT_TYPE, messageToJson(BAD_GATEWAY, -1))
			return
		}
		defer r.Body.Close()

		err = json.Unmarshal([]byte(lines), &cars)
		if err != nil {
			respond(w, 502, CONTENT_TYPE, messageToJson(BAD_GATEWAY, -1))
			return
		}

		respond(w, 202, CONTENT_TYPE, messageToJson(ACCEPTED, len(cars)))
	}
}

// Indicate the service has started up correctly and is ready to accept requests.
func status(w http.ResponseWriter, r *http.Request) {
	rsp, err := http.Get(IP_ADDRESS + strconv.Itoa(PORT))
	if err != nil {
		log.Fatal(err)
	}

	body, _ := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()

	if rsp.StatusCode == 200 {
		respond(w, 200, CONTENT_TYPE, messageToJson(OK, -1))
	} else {
		w.Write([]byte(body))
	}
}

// Add a group.
func group(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		groupCount++
		newGroup := Group{Id: groupCount}
		groups = append(groups, newGroup)

		respond(w, 200, CONTENT_TYPE, messageToJson(OK, groupCount))
	}
}

// A group of people requests to perform a journey.
func journey(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		groupId, gerr := strconv.Atoi(r.FormValue("gid"))
		seats, serr := strconv.Atoi(r.FormValue("seats"))

		if gerr != nil || serr != nil {
			respond(w, 400, CONTENT_TYPE, messageToJson(BAD_REQUEST, -1))
			return
		}

		// Group ID check.
		if !findGroup(groups, len(groups), groupId) {
			respond(w, 404, CONTENT_TYPE, messageToJson(NOT_FOUND, -1))
			return
		}

		// Find a car.
		carId := findCarBySeats(cars, len(cars), seats)
		journeyCount++
		newJourney := Journey{Id: journeyCount, GroupId: groupId, CarId: carId}
		journeys = append(journeys, newJourney)

		// Mark car as engaged.
		rawCarId := findCarById(cars, len(cars), carId)
		if rawCarId >= 0 {
			cars[rawCarId].Engaged = true
		}

		respond(w, 200, CONTENT_TYPE, messageToJson(OK, journeyCount))
	}
}

// Return the car the group is traveling with, or no car if they are still waiting to be served.
func locate(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		var carId int

		groupId, err := strconv.Atoi(r.FormValue("gid"))
		if err != nil {
			respond(w, 400, CONTENT_TYPE, messageToJson(BAD_REQUEST, -1))
			return
		}

		// Group ID check.
		if !findGroup(groups, len(groups), groupId) {
			respond(w, 404, CONTENT_TYPE, messageToJson(NOT_FOUND, -1))
			return
		}

		journeyId := findJourneyByJid(journeys, len(journeys), groupId)
		if journeyId >= len(journeys) {
			respond(w, 502, CONTENT_TYPE, messageToJson(BAD_GATEWAY, journeyId))
			return
		}

		if len(journeys) > 0 && journeyId > -1 {
			carId = journeys[journeyId].CarId
		} else {
			respond(w, 400, CONTENT_TYPE, messageToJson(BAD_REQUEST, journeyId))
			return
		}

		if carId <= 0 {
			respond(w, 404, CONTENT_TYPE, messageToJson(NOT_FOUND, -1))
			return
		} else if carId > 0 {
			respond(w, 200, CONTENT_TYPE, messageToJson(OK, carId))
			return
		}
	}
}

// A group of people requests to be dropped off.
func dropOff(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		groupId, err := strconv.Atoi(r.FormValue("gid"))
		if err != nil {
			respond(w, 400, CONTENT_TYPE, messageToJson(BAD_REQUEST, -1))
			return
		}

		// Group ID check.
		if !findGroup(groups, len(groups), groupId) {
			respond(w, 404, CONTENT_TYPE, messageToJson(NOT_FOUND, -1))
			return
		}

		journeyId := findJourneyByJid(journeys, len(journeys), groupId)
		groupId = findGroupByGid(groups, len(groups), groupId)

		if journeyId == -1 {
			respond(w, 404, CONTENT_TYPE, messageToJson(NOT_FOUND, -1))
			return
		} else {
			journeys = append(journeys[:journeyId], journeys[journeyId+1:]...)
			groups = append(groups[:groupId], groups[groupId+1:]...)
		}

		respond(w, 200, CONTENT_TYPE, messageToJson(OK, -1))
	}
}

func onRequest() {
	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/", home)
	router.HandleFunc("/status", status)
	router.HandleFunc("/group", group)
	router.HandleFunc("/journey", journey)
	router.HandleFunc("/locate", locate)
	router.HandleFunc("/dropoff", dropOff)
	router.HandleFunc("/cars", loadCars)

	log.Fatal(http.ListenAndServe(":12345", router))
}

func main() {
	groups = make([]Group, 0)

	cars = make([]Car, 0)
	cars = append(cars, Car{Id: 1, Seats: 5})
	cars = append(cars, Car{Id: 4, Seats: 4})
	cars = append(cars, Car{Id: 3, Seats: 2})
	cars = append(cars, Car{Id: 2, Seats: 6})

	journeys = make([]Journey, 0)

	fmt.Println("Car Availability Service up and running on port " + strconv.Itoa(PORT))
	onRequest()
}
