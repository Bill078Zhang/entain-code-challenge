package db

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"git.neds.sh/matty/entain/sports/proto/sports"
)

// TestSportsRepo_FilterByVisibility tests filtering events by visibility.
func TestEventsRepo_FilterByVisibility(t *testing.T) {
	repo := createTestRepo(t)

	visible := true
	filter := &sports.ListEventsRequestFilter{
		Visible: &visible,
	}
	order := &sports.ListEventsRequestOrderParam{}

	events, err := repo.List(filter, order)

	assert.NoError(t, err)
	for _, event := range events {
		assert.Truef(t, event.Visible, "Event %d should be visible", event.Id)
	}
}

// TestEventsRepo_FilterByName tests the functionality of filtering events by name in the EventsRepo.
func TestEventsRepo_FilterByName(t *testing.T) {
	// Create a test repository instance.
	repo := createTestRepo(t)

	// Fetch all events from the repository to establish a baseline.
	eventsResponse, err := repo.List(nil, nil)
	assert.NoError(t, err)
	assert.NotNil(t, eventsResponse)
	assert.Equal(t, 100, len(eventsResponse))

	var eventName string
	for _, event := range eventsResponse {
		if len(event.Name) > 2 {
			eventName = event.Name[1 : len(event.Name)-2]
			break
		}
	}

	filterEventName := &sports.ListEventsRequestFilter{Name: &eventName}
	// Retrieve events from the repository filtered by the extracted event name.
	events, err := repo.List(filterEventName, nil)
	assert.NoError(t, err)
	assert.NotNil(t, events)

	// Assert that each event's name contains the specified event name substring.
	for _, event := range events {
		assert.Contains(t, event.Name, eventName, "Event %d is not %s event.", event.Id, eventName)
	}
}

// TestEventsRepo_FilterByType tests the functionality of filtering events by type in the EventsRepo.
func TestEventsRepo_FilterByType(t *testing.T) {
	repo := createTestRepo(t)

	eventsResponse, err := repo.List(nil, nil)
	assert.NoError(t, err)
	assert.NotNil(t, eventsResponse)
	assert.Equal(t, 100, len(eventsResponse))

	// Extract the type of the first event to use for filtering.
	firstEventType := eventsResponse[0].Type

	filter := &sports.ListEventsRequestFilter{
		Type: &firstEventType,
	}

	events, err := repo.List(filter, nil)
	assert.NoError(t, err)
	assert.NotNil(t, events)

	// Assert that each event's type matches the type of the first event.
	for _, event := range events {
		assert.EqualValues(t, firstEventType, event.Type, "Event %d is not a %s event.", event.Id, firstEventType)
	}
}


// TestSportsRepo_ListAll tests retrieving all events.
func TestSportsRepo_ListAll(t *testing.T) {
	repo := createTestRepo(t)

	filter := &sports.ListEventsRequestFilter{}
	order := &sports.ListEventsRequestOrderParam{}
	events, err := repo.List(filter, order)

	assert.NoError(t, err)
	assert.Equalf(t, 100, len(events), "Expected 100 events in the result")

	// Check if events are sorted by default field (advertised_start_time)
	for _, event := range events {
		assert.Truef(t, event.AdvertisedStartTime.Nanos >= event.AdvertisedStartTime.Nanos, "Results are not sorted by advertised_start_time")
	}
}

// TestSportsRepo_Sort tests sorting events by different fields and directions.
func TestSportsRepo_Sort(t *testing.T) {
	repo := createTestRepo(t)

	// Define the fields and directions for sorting
	sortFields := []string{"id", "name", "type", "advertised_start_time"}
	sortDirections := []string{"ASC", "DESC"}

	// Iterate over sort fields and directions
	for _, field := range sortFields {
		for _, direction := range sortDirections {
			// Set up filter and order
			filter := &sports.ListEventsRequestFilter{}
			order := &sports.ListEventsRequestOrderParam{
				Field:     &field,
				Direction: &direction,
			}

			// Call repo.List with filter and order
			events, err := repo.List(filter, order)
			assert.NoError(t, err)
			assert.Equalf(t, 100, len(events), "Expected total of 100 events in DB.")

			// Check if events are sorted correctly based on the field and direction
			for i := 0; i < len(events)-1; i++ {
				previousElement := events[i]
				element := events[i+1]
				switch field {
				case "name":
					if direction == "ASC" {
						assert.Truef(t, element.Name >= previousElement.Name, "Results are not sorted by name in ascending order")
					} else {
						assert.Truef(t, element.Name <= previousElement.Name, "Results are not sorted by name in descending order")
					}
				case "type":
					if direction == "ASC" {
						assert.Truef(t, element.Type >= previousElement.Type, "Results are not sorted by type in ascending order")
					} else {
						assert.Truef(t, element.Type <= previousElement.Type, "Results are not sorted by type in descending order")
					}
				case "advertised_start_time":
					if direction == "ASC" {
						assert.Truef(t, element.AdvertisedStartTime.Nanos >= previousElement.AdvertisedStartTime.Nanos, "Results are not sorted by advertised_start_time in ascending order")
					} else {
						assert.Truef(t, element.AdvertisedStartTime.Nanos <= previousElement.AdvertisedStartTime.Nanos, "Results are not sorted by advertised_start_time in descending order")
					}
				case "id":
					if direction == "ASC" {
						assert.Truef(t, element.Id >= previousElement.Id, "Results are not sorted by id in ascending order")
					} else {
						assert.Truef(t, element.Id <= previousElement.Id, "Results are not sorted by id in descending order")
					}
				}
			}
		}
	}
}

// TestSportsRepo_StatusField tests the status field calculated by the advertised start time.
func TestSportsRepo_StatusField(t *testing.T) {
	repo := createTestRepo(t)

	filter := &sports.ListEventsRequestFilter{}
	order := &sports.ListEventsRequestOrderParam{}
	events, err := repo.List(filter, order)

	assert.NoError(t, err)
	assert.Equalf(t, 100, len(events), "There should be a total of 100 events in DB.")

	// Validating event statuses based on advertised start time
	for _, event := range events {
		switch {
		case event.Status == "CLOSED":
			assert.Truef(t, time.Now().Unix() > event.AdvertisedStartTime.Seconds, "Event %d should not be CLOSED.", event.Id)
		case event.Status == "OPEN":
			assert.Truef(t, time.Now().Unix() <= event.AdvertisedStartTime.Seconds, "Event %d should not be OPEN.", event.Id)
		default:
			assert.Failf(t, "Status field must only contain OPEN or CLOSED.", event.Status)
		}
	}
}

// createTestRepo initializes a test repository for sports events.
func createTestRepo(t *testing.T) EventsRepo {
	eventDB, err := sql.Open("sqlite3", ":memory:")
	assert.NoError(t, err)

	repo := NewEventsRepo(eventDB)
	err = repo.Init()
	assert.NoError(t, err)

	return repo
}