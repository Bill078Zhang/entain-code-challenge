package db

import "fmt"

const (
	eventsList = "list"
)

// DB constants with table sports field names
const (
	Id                  = "id"
	Name                = "name"
	Type               	= "type"
	Location            = "location"
	Visible			 	= "visible"
	AdvertisedStartTime = "advertised_start_time"
)

func getEventQueries() map[string]string {
	return map[string]string{
		eventsList: fmt.Sprintf(
			"SELECT %s, %s, %s, %s, %s, %s FROM sport_events",
			Id,
			Name,
			Type,
			Location,
			Visible,
			AdvertisedStartTime),
	}
}
