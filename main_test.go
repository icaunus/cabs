package main

import "testing"

func TestFindCar(t *testing.T) {
	testCars := []Car{{Id: 1, Seats: 4}, {Id: 0, Seats: 5}, {Id: 2, Seats: 6}, {Id: 3, Seats: 7}}
	const FAKE_SEATS = 9
	const REAL_SEATS = 4

	fakeCarId := findCarBySeats(testCars, len(testCars), FAKE_SEATS)
	if fakeCarId != -1 {
		t.Error("Non-existent car found: ", fakeCarId)
	}

	realCarId := findCarBySeats(testCars, len(testCars), REAL_SEATS)
	if realCarId != 1 {
		t.Error("Existing car NOT found: ", realCarId)
	}
}

func TestFindGroup(t *testing.T) {
	const FAKE_GROUP_ID = 1234
	const REAL_GROUP_ID = 111
	testGroups := []Group{{Id: 0}, {Id: 111}, {Id: 3333}, {Id: 44}}

	fakeGroupFound := findGroup(testGroups, len(testGroups), FAKE_GROUP_ID)
	if fakeGroupFound {
		t.Error("Non-exsistent group found: ", FAKE_GROUP_ID)
	}

	realGroupFound := findGroup(testGroups, len(testGroups), REAL_GROUP_ID)
	if !realGroupFound {
		t.Error("Existing group NOT found: ", REAL_GROUP_ID)
	}
}

func TestFindJourneyByGroup(t *testing.T) {
	const FAKE_GROUP_ID = 1234
	const REAL_GROUP_ID = 3
	const REAL_JOURNEY_ID = 2
	var journeyId int

	testJourneys := []Journey{{Id: 4, GroupId: 1, CarId: 11}, {Id: 3, GroupId: 2, CarId: 22}, {Id: 2, GroupId: 3, CarId: 33}, {Id: 1, GroupId: 4, CarId: 44}}

	journeyId = findJourneyByGroup(testJourneys, len(testJourneys), FAKE_GROUP_ID)
	if journeyId != -1 {
		t.Error("Fake journey found: ", journeyId)
	}

	journeyId = findJourneyByGroup(testJourneys, len(testJourneys), REAL_GROUP_ID)
	if journeyId != REAL_JOURNEY_ID {
		t.Error("Existing journey NOT found: ", REAL_JOURNEY_ID)
	}
}
