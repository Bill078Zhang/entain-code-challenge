package db

import (
	"database/sql"
	"strings"
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes"
	_ "github.com/mattn/go-sqlite3"

	"git.neds.sh/matty/entain/sports/proto/sports"
)

// ValidColumns defines a slice of valid column names for ordering.
var ValidColumns = []string{"id", "name", "type", "advertised_start_time"}

// EventsRepo provides repository access to sport_events.
type EventsRepo interface {
	// Init will initialise our sports repository.
	Init() error

	// List will return a list of events.
	List(filter *sports.ListEventsRequestFilter, order *sports.ListEventsRequestOrderParam) ([]*sports.Event, error)
}

type eventsRepo struct {
	db   *sql.DB
	init sync.Once
}

// NewEventsRepo creates a new events repository.
func NewEventsRepo(db *sql.DB) EventsRepo {
	return &eventsRepo{db: db}
}

// Init prepares the event repository dummy data.
func (r *eventsRepo) Init() error {
	var err error

	r.init.Do(func() {
		// For test/example purposes, we seed the DB with some dummy sports.
		err = r.seed()
	})

	return err
}

func (r *eventsRepo) List(filter *sports.ListEventsRequestFilter, order *sports.ListEventsRequestOrderParam) ([]*sports.Event, error) {
	var (
		err   error
		query string
		args  []interface{}
	)

	query = getEventQueries()[eventsList]

	query, args = r.applyFilter(query, filter)

	query = r.applyOrder(query, order)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}

	return r.scanEvents(rows)
}

func (r *eventsRepo) applyFilter(query string, filter *sports.ListEventsRequestFilter) (string, []interface{}) {
	var (
		clauses []string
		args    []interface{}
	)

	if filter == nil {
		return query, args
	}


	if filter.Name != nil {
		clauses = append(clauses, "name LIKE ?")
		args = append(args, "%"+*filter.Name+"%")
	}

	if filter.Type != nil {
		clauses = append(clauses, "type = ?")
		args = append(args, filter.Type)
	}

	if filter.Location != nil {
		clauses = append(clauses, "location = ?")
		args = append(args, filter.Location)
	}

	if filter.Visible != nil {
		clauses = append(clauses, "visible = ?")
		args = append(args, filter.Visible)
	}

	if len(clauses) != 0 {
		query += " WHERE " + strings.Join(clauses, " AND ")
	}

	return query, args
}

// applyOrder applies ordering to the query based on the user-provided field and direction.
func (r *eventsRepo) applyOrder(query string, order *sports.ListEventsRequestOrderParam) string {
	// Return immediately if no order is specified
	if order == nil {
		return query
	}

	// Default order by if no field has been provided
	if order.Field == nil {
		query += " ORDER BY advertised_start_time"
	} else {
		// Validate user-provided field
		if !isValidColumn(*order.Field) {
			return query
		}
		query += " ORDER BY " + *order.Field
	}

	// Append direction if provided and valid
	if order.Direction != nil && (strings.EqualFold(*order.Direction, "ASC") || strings.EqualFold(*order.Direction, "DESC")) {
		query += " " + strings.ToUpper(*order.Direction)
	}	

	return query
}

// isValidColumn checks if the provided column name is valid for ordering.
func isValidColumn(column string) bool {
	for _, validColumn := range ValidColumns {
		if strings.EqualFold(column, validColumn) {
			return true
		}
	}
	return false
}

func (r *eventsRepo) scanEvents(
	rows *sql.Rows,
) ([]*sports.Event, error) {
	var events []*sports.Event

	for rows.Next() {
		var event sports.Event
		var advertisedStart time.Time

		if err := rows.Scan(&event.Id, &event.Name, &event.Type, &event.Location, &event.Visible, &advertisedStart); err != nil {
			if err == sql.ErrNoRows {
				return nil, nil
			}

			return nil, err
		}

		ts, err := ptypes.TimestampProto(advertisedStart)
		if err != nil {
			return nil, err
		}

		event.AdvertisedStartTime = ts

		// Calculates the status field value based on a race's advertisedStartTime.
		event.Status = "OPEN"
		if ptypes.TimestampNow().Seconds > ts.Seconds {
			event.Status = "CLOSED"
		}

		events = append(events, &event)
	}

	return events, nil
}
