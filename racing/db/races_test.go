package db

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"git.neds.sh/matty/entain/racing/proto/racing"
)

// TestRacesRepoWithMeetingIDs_FilteredList tests filtering races by specific meeting IDs.
func TestRacesRepo_FilteredByMeetingIds(t *testing.T) {
    repo := createTestRepo(t)
    meetingIDs := []int64{2, 3, 5, 7, 11, 13}

    filter := &racing.ListRacesRequestFilter{
        MeetingIds: meetingIDs,
    }
	order := &racing.ListRacesRequestOrderParam{}
    races, err := repo.List(filter, order)

    assert.NoError(t, err)
    // Ensure that all returned races have one of the specified meeting IDs
    for _, race := range races {
        assert.Contains(t, meetingIDs, race.MeetingId, "Race with unexpected meeting ID found.")
    }
}

// TestRacesRepo_FilterByVisibility tests filtering races by visibility.
func TestRacesRepo_FilterByVisibility(t *testing.T) {
	repo := createTestRepo(t)

	visible := true
	filter := &racing.ListRacesRequestFilter{
		Visible: &visible,
	}
	order := &racing.ListRacesRequestOrderParam{}
	races, err := repo.List(filter, order)

	assert.NoError(t, err)
	for _, race := range races {
		assert.Truef(t, race.Visible, "Race %d should be visible", race.Id)
	}
}

// TestRacesRepo_ListAll tests retrieving all races.
func TestRacesRepo_ListAll(t *testing.T) {
	repo := createTestRepo(t)

	filter := &racing.ListRacesRequestFilter{}
	order := &racing.ListRacesRequestOrderParam{}
	races, err := repo.List(filter, order)

	assert.NoError(t, err)
	assert.Equalf(t, 100, len(races), "Expected 100 races in the result")

	// Check if races are sorted by default field (advertised_start_time)
	for _, race := range races {
		assert.Truef(t, race.AdvertisedStartTime.Nanos >= race.AdvertisedStartTime.Nanos, "Results are not sorted by advertised_start_time")
	}
}

// TestRacesRepo_Sort tests sorting races by different fields and directions.
func TestRacesRepo_Sort(t *testing.T) {
	repo := createTestRepo(t)

	// Define the fields and directions for sorting
	sortFields := []string{"id", "meeting_id", "name", "advertised_start_time"}
	sortDirections := []string{"ASC", "DESC"}

	// Iterate over sort fields and directions
	for _, field := range sortFields {
		for _, direction := range sortDirections {
			// Set up filter and order
			filter := &racing.ListRacesRequestFilter{}
			order := &racing.ListRacesRequestOrderParam{
				Field:     &field,
				Direction: &direction,
			}

			// Call repo.List with filter and order
			races, err := repo.List(filter, order)
			assert.NoError(t, err)
			assert.Equalf(t, 100, len(races), "Expected total of 100 races in DB.")

			// Check if races are sorted correctly based on the field and direction
			for i := 0; i < len(races)-1; i++ {
				previousElement := races[i]
				element := races[i+1]
				switch field {
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

// createTestRepo initializes a test repository for races.
func createTestRepo(t *testing.T) RacesRepo {
	racingDB, err := sql.Open("sqlite3", ":memory:")
	assert.NoError(t, err)

	repo := NewRacesRepo(racingDB)
	err = repo.Init()
	assert.NoError(t, err)

	return repo
}
