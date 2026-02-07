package models

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type Dataset struct {
	Airports []Airport `json:"airports"`
	Flights  []Flight  `json:"flights"`
}

type Airport struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	City     string `json:"city"`
	Country  string `json:"country"`
	Timezone string `json:"timezone"`
}

type Flight struct {
	FlightNumber  string `json:"flightNumber"`
	Airline       string `json:"airline"`
	Origin        string `json:"origin"`
	Destination   string `json:"destination"`
	DepartureTime string `json:"departureTime"`
	ArrivalTime   string `json:"arrivalTime"`
	Price         Price  `json:"price"`
	Aircraft      string `json:"aircraft"`
}

type Price float64

func (p *Price) UnmarshalJSON(b []byte) error {
	var num float64
	if err := json.Unmarshal(b, &num); err == nil {
		*p = Price(num)
		return nil
	}

	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return fmt.Errorf("invalid price string %q: %w", s, err)
		}
		*p = Price(f)
		return nil
	}

	return fmt.Errorf("invalid price: %s", string(b))
}

// API output models
type SegmentOut struct {
	FlightNumber   string  `json:"flightNumber"`
	Airline        string  `json:"airline"`
	Origin         string  `json:"origin"`
	Destination    string  `json:"destination"`
	DepartureLocal string  `json:"departureLocal"`
	ArrivalLocal   string  `json:"arrivalLocal"`
	Price          float64 `json:"price"`
	Aircraft       string  `json:"aircraft"`
}

type ItineraryOut struct {
	Segments             []SegmentOut `json:"segments"`
	LayoversMinutes      []int        `json:"layoversMinutes"`
	TotalDurationMinutes int          `json:"totalDurationMinutes"`
	TotalPrice           float64      `json:"totalPrice"`
}

type SearchResponse struct {
	Origin      string         `json:"origin"`
	Destination string         `json:"destination"`
	Date        string         `json:"date"`
	Count       int            `json:"count"`
	Itineraries []ItineraryOut `json:"itineraries"`
}

type SearchError struct {
	Message string `json:"message"`
}
