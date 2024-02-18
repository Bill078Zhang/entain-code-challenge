package service

import (
	"database/sql"
	"testing"

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
