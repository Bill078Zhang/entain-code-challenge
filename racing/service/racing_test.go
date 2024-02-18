package service

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"git.neds.sh/matty/entain/racing/db"
	"git.neds.sh/matty/entain/racing/proto/racing"
)

func TestRacingService_FilteredByMeetingIds(t *testing.T) {
	racingService := createService(t)

	meetingIDs := []int64{1, 2}
	filter := &racing.ListRacesRequestFilter{
		MeetingIds: meetingIDs,
	}
	request := &racing.ListRacesRequest{Filter: filter}

	response, err := racingService.ListRaces(nil, request)

	assert.NoError(t, err)
	for _, race := range response.Races {
		assert.Contains(t, meetingIDs, race.MeetingId, "Race with unexpected meeting ID found.")
	}
}

func TestRacingService_FilteredByVisibility(t *testing.T) {
	racingService := createService(t)

	filter := &racing.ListRacesRequestFilter{
		Visible: BoolPtr(true),
	}
	request := &racing.ListRacesRequest{Filter: filter}

	response, err := racingService.ListRaces(nil, request)

	assert.NoError(t, err)
	for _, race := range response.Races {
		assert.Truef(t, race.Visible, "Race %d is not visible.", race.Id)
	}
}

func TestRacingService_GetAllRaces(t *testing.T) {
	racingService := createService(t)

	request := &racing.ListRacesRequest{Filter: &racing.ListRacesRequestFilter{}}

	response, err := racingService.ListRaces(nil, request)

	assert.NoError(t, err)
	assert.Equalf(t, 100, len(response.Races), "There should be a total of 100 races in DB.")
}

func TestRacingService_SortFields(t *testing.T) {
	racingService := createService(t)
	sortFields := []string{"id", "meeting_id", "name", "advertised_start_time"}

	// Iterate over sort fields
	for _, field := range sortFields {
		// Iterate over sort directions
		for _, direction := range []string{"ASC", "DESC"} {
			order := &racing.ListRacesRequestOrderParam{
				Field:     &field,
				Direction: &direction,
			}

			listRacesRequest := &racing.ListRacesRequest{Order: order}
			listRacesResponse, err := racingService.ListRaces(nil, listRacesRequest)
			assert.NoError(t, err)
			assert.Equalf(t, 100, len(listRacesResponse.Races), "There should be a total of 100 races in DB.")

			// Check if races are sorted correctly based on the field and direction
			for i := 0; i < len(listRacesResponse.Races)-1; i++ {
				previousElement := listRacesResponse.Races[i]
				element := listRacesResponse.Races[i+1]
				switch field {
				case "id":
					if direction == "ASC" {
						assert.Truef(t, element.Id >= previousElement.Id, "Results are not sorted by id in ascending order")
					} else {
						assert.Truef(t, element.Id <= previousElement.Id, "Results are not sorted by id in descending order")
					}
				case "meeting_id":
					if direction == "ASC" {
						assert.Truef(t, element.MeetingId >= previousElement.MeetingId, "Results are not sorted by meeting_id in ascending order")
					} else {
						assert.Truef(t, element.MeetingId <= previousElement.MeetingId, "Results are not sorted by meeting_id in descending order")
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

func TestRacingService_SortByDefault(t *testing.T) {
	racingService := createService(t)

	listRacesRequest := &racing.ListRacesRequest{}
	listRacesResponse, err := racingService.ListRaces(nil, listRacesRequest)
	assert.NoError(t, err)
	assert.Equalf(t, 100, len(listRacesResponse.Races), "There should be a total of 100 races in DB.")
	
	// Check if races are sorted by default field (advertised_start_time)
	for i := 0; i < len(listRacesResponse.Races)-1; i++ {
		previousElement := listRacesResponse.Races[i]
		element := listRacesResponse.Races[i+1]
		assert.Truef(t, element.AdvertisedStartTime.Nanos >= previousElement.AdvertisedStartTime.Nanos, "Results are not sorted by advertised_start_time")
	}
}

func TestRacingService_StatusField(t *testing.T) {
	racingService := createService(t)

	listRacesRequest := &racing.ListRacesRequest{Filter: &racing.ListRacesRequestFilter{}}
	listRacesResponse, err := racingService.ListRaces(nil, listRacesRequest)

	assert.NoError(t, err)
	assert.Equalf(t, 100, len(listRacesResponse.Races), "There should be a total of 100 races in DB.")

	// Validating race statuses based on advertised start time
	for _, race := range listRacesResponse.Races {
		switch {
		case race.Status == "CLOSED":
			assert.Truef(t, time.Now().Unix() > race.AdvertisedStartTime.Seconds, "Race %d should not be CLOSED.", race.Id)
		case race.Status == "OPEN":
			assert.Truef(t, time.Now().Unix() <= race.AdvertisedStartTime.Seconds, "Race %d should not be OPEN.", race.Id)
		default:
			assert.Failf(t, "Status field must only contain OPEN or CLOSED.", race.Status)
		}
	}
}

func TestRacingService_GetRace(t *testing.T) {
	racingService := createService(t)

	getRaceRequest := &racing.GetRaceRequest{Id: 35}
	race, err := racingService.GetRace(nil, getRaceRequest)

    assert.NoError(t, err, "No error should occur.")
    assert.NotNil(t, race, "Race should exist.")
    assert.EqualValuesf(t, 35, race.Id, "Race with ID 35 should exist.")
}

func TestRacingService_GetRaceByInvalidId(t *testing.T) {
	racingService := createService(t)

	getRaceRequest := &racing.GetRaceRequest{Id: -1}
	race, err := racingService.GetRace(nil, getRaceRequest)

    assert.Error(t, err)
    assert.Nil(t, race, "Race should not exist.")
}

func createService(t *testing.T) Racing {
	racingDB, err := sql.Open("sqlite3", ":memory:")
	assert.NoError(t, err)

	// Initialize the races repository
	repo := db.NewRacesRepo(racingDB)
	err = repo.Init()
	assert.NoError(t, err)

	return NewRacingService(repo)
}

func BoolPtr(b bool) *bool {
	return &b
}
