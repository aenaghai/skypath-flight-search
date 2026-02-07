package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

type Store struct {
	AirportsByCode  map[string]Airport
	FlightsByOrigin map[string][]Flight // origin -> flight
}

func LoadStore(path string) (*Store, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var ds Dataset
	if err := json.Unmarshal(b, &ds); err != nil {
		return nil, err
	}

	air := make(map[string]Airport, len(ds.Airports))
	for _, a := range ds.Airports {
		air[strings.ToUpper(a.Code)] = a
	}

	flightBy := make(map[string][]Flight)
	for _, f := range ds.Flights {
		f.Origin = strings.ToUpper(f.Origin)
		f.Destination = strings.ToUpper(f.Destination)
		flightBy[f.Origin] = append(flightBy[f.Origin], f)
	}

	return &Store{
		AirportsByCode:  air,
		FlightsByOrigin: flightBy,
	}, nil
}

func (s *Store) GetAirport(code string) (Airport, bool) {
	a, ok := s.AirportsByCode[strings.ToUpper(code)]
	return a, ok
}

func (s *Store) Country(code string) (string, bool) {
	a, ok := s.GetAirport(code)
	if !ok {
		return "", false
	}
	return a.Country, true
}

func (s *Store) TZLocation(airportCode string) (*time.Location, error) {
	a, ok := s.GetAirport(airportCode)
	if !ok {
		return nil, fmt.Errorf("unknown airport : %s", airportCode)
	}

	loc, err := time.LoadLocation(a.Timezone)
	if err != nil {
		return nil, fmt.Errorf("invalid timezone for airport %s: %w", airportCode, err)
	}
	return loc, nil
}
