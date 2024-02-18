package service

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"git.neds.sh/matty/entain/sports/db"
	"git.neds.sh/matty/entain/sports/proto/sports"
)

func TestSportsService_FilteredByVisibility(t *testing.T) {
	sportsService := createService(t)

	filter := &sports.ListEventsRequestFilter{
		Visible: BoolPtr(true),
	}
	request := &sports.ListEventsRequest{Filter: filter}

	response, err := sportsService.ListEvents(nil, request)

	assert.NoError(t, err)
	for _, event := range response.Events {
		assert.Truef(t, event.Visible, "Event %d is not visible.", event.Id)
	}
}

func TestSportsService_FilteredByName(t *testing.T) {
	sportsService := createService(t)

	listEventsRequest := &sports.ListEventsRequest{Filter: &sports.ListEventsRequestFilter{}}
	listEventsResponse, err := sportsService.ListEvents(nil, listEventsRequest)
	assert.NoError(t, err)
	assert.NotNil(t, listEventsResponse)
	assert.Equal(t, 100, len(listEventsResponse.Events))

	var eventName string
	for i := 0; i < 100; i++ {
		event := listEventsResponse.Events[i].Name
		if len(event) > 2 {
			eventName = event[1 : len(event)-2]
			break
		}
	}

	filter := &sports.ListEventsRequestFilter{Name: &eventName}
	listEventsRequestByName := &sports.ListEventsRequest{Filter: filter}
	listEventsResponseByName, err := sportsService.ListEvents(nil, listEventsRequestByName)
	assert.NoError(t, err)
	assert.NotNil(t, listEventsResponseByName)

	// Assert that each event's name contains the specified event name substring.
	for _, event := range listEventsResponseByName.Events {
		assert.Contains(t, event.Name, eventName, "Event %d does not match the expected name substring %s.", event.Id, eventName)
	}
}

func TestSportsService_FilteredByType(t *testing.T) {
	sportsService := createService(t)

	listEventsRequest := &sports.ListEventsRequest{Filter: &sports.ListEventsRequestFilter{}}
	listEventsResponse, err := sportsService.ListEvents(nil, listEventsRequest)
	assert.NoError(t, err)
	assert.NotNil(t, listEventsResponse)
	assert.Equal(t, 100, len(listEventsResponse.Events))

	// Extract the type of the first event to use for filtering.
	firstEventType := listEventsResponse.Events[0].Type

	filter := &sports.ListEventsRequestFilter{Type: &firstEventType}
	listEventsRequestByType := &sports.ListEventsRequest{Filter: filter}
	listEventsResponseByType, err := sportsService.ListEvents(nil, listEventsRequestByType)
	assert.NoError(t, err)
	assert.NotNil(t, listEventsResponseByType)

	// Assert that each event's type matches the type of the first event.
	for _, event := range listEventsResponseByType.Events {
		assert.EqualValues(t, firstEventType, event.Type, "Event %d is not of the expected type %s.", event.Id, firstEventType)
	}
}

func TestSportsService_FetchAllEvents(t *testing.T) {
	sportsService := createService(t)

	request := &sports.ListEventsRequest{Filter: &sports.ListEventsRequestFilter{}}

	response, err := sportsService.ListEvents(nil, request)

	assert.NoError(t, err)
	assert.Equalf(t, 100, len(response.Events), "There should be a total of 100 events in DB.")
}

func TestSportsService_SortFields(t *testing.T) {
	sportsService := createService(t)
	sortFields := []string{"id", "type", "name", "advertised_start_time"}

	// Iterate over sort fields
	for _, field := range sortFields {
		// Iterate over sort directions
		for _, direction := range []string{"ASC", "DESC"} {
			order := &sports.ListEventsRequestOrderParam{
				Field:     &field,
				Direction: &direction,
			}

			listEventsRequest := &sports.ListEventsRequest{Order: order}
			listEventsResponse, err := sportsService.ListEvents(nil, listEventsRequest)
			assert.NoError(t, err)
			assert.Equalf(t, 100, len(listEventsResponse.Events), "There should be a total of 100 events in DB.")

			// Check if events are sorted correctly based on the field and direction
			for i := 0; i < len(listEventsResponse.Events)-1; i++ {
				previousElement := listEventsResponse.Events[i]
				element := listEventsResponse.Events[i+1]
				switch field {
				case "id":
					if direction == "ASC" {
						assert.Truef(t, element.Id >= previousElement.Id, "Results are not sorted by id in ascending order")
					} else {
						assert.Truef(t, element.Id <= previousElement.Id, "Results are not sorted by id in descending order")
					}
				case "type":
					if direction == "ASC" {
						assert.Truef(t, element.Type >= previousElement.Type, "Results are not sorted by type in ascending order")
					} else {
						assert.Truef(t, element.Type <= previousElement.Type, "Results are not sorted by type in descending order")
					}
				case "name":
					if direction == "ASC" {
						assert.Truef(t, element.Name >= previousElement.Name, "Results are not sorted by name in ascending order")
					} else {
						assert.Truef(t, element.Name <= previousElement.Name, "Results are not sorted by name in descending order")
					}
				case "advertised_start_time":
					if direction == "ASC" {
						assert.Truef(t, element.AdvertisedStartTime.Nanos >= previousElement.AdvertisedStartTime.Nanos, "Results are not sorted by advertised_start_time in ascending order")
					} else {
						assert.Truef(t, element.AdvertisedStartTime.Nanos <= previousElement.AdvertisedStartTime.Nanos, "Results are not sorted by advertised_start_time in descending order")
					}
				}
			}
		}
	}
}

func TestSportsService_SortByDefault(t *testing.T) {
	sportsService := createService(t)

	listEventsRequest := &sports.ListEventsRequest{}
	listEventsResponse, err := sportsService.ListEvents(nil, listEventsRequest)
	assert.NoError(t, err)
	assert.Equalf(t, 100, len(listEventsResponse.Events), "There should be a total of 100 events in DB.")
	
	// Check if events are sorted by default field (advertised_start_time)
	for i := 0; i < len(listEventsResponse.Events)-1; i++ {
		previousElement := listEventsResponse.Events[i]
		element := listEventsResponse.Events[i+1]
		assert.Truef(t, element.AdvertisedStartTime.Nanos >= previousElement.AdvertisedStartTime.Nanos, "Results are not sorted by advertised_start_time")
	}
}

func TestSportsService_StatusField(t *testing.T) {
	sportsService := createService(t)

	listEventsRequest := &sports.ListEventsRequest{Filter: &sports.ListEventsRequestFilter{}}
	listEventsResponse, err := sportsService.ListEvents(nil, listEventsRequest)

	assert.NoError(t, err)
	assert.Equalf(t, 100, len(listEventsResponse.Events), "There should be a total of 100 events in DB.")

	// Validating event statuses based on advertised start time
	for _, event := range listEventsResponse.Events {
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

func createService(t *testing.T) Sports {
	eventDB, err := sql.Open("sqlite3", ":memory:")
	assert.NoError(t, err)

	// Initialize the events repository
	repo := db.NewEventsRepo(eventDB)
	err = repo.Init()
	assert.NoError(t, err)

	return NewSportsService(repo)
}

func BoolPtr(b bool) *bool {
	return &b
}