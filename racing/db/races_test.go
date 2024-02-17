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
    races, err := repo.List(filter)

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
	races, err := repo.List(filter)

	assert.NoError(t, err)
	for _, race := range races {
		assert.Truef(t, race.Visible, "Race %d should be visible", race.Id)
	}
}

// TestRacesRepo_ListAll tests retrieving all races.
func TestRacesRepo_ListAll(t *testing.T) {
	repo := createTestRepo(t)

	filter := &racing.ListRacesRequestFilter{}
	races, err := repo.List(filter)

	assert.NoError(t, err)
	assert.Equalf(t, 100, len(races), "Expected 100 races in the result")
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
