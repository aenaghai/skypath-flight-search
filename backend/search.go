package main

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

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

func SearchItineraries(s *Store, origin, destination, dateStr string) ([]ItineraryOut, error) {
	origin = strings.ToUpper(strings.TrimSpace(origin))
	destination = strings.ToUpper(strings.TrimSpace(destination))

	if origin == destination {
		return []ItineraryOut{}, nil
	}
	if _, ok := s.GetAirport(origin); !ok {
		return nil, fmt.Errorf("invalid origin airport code: %s", origin)
	}
	if _, ok := s.GetAirport(destination); !ok {
		return nil, fmt.Errorf("invalid destination airport code: %s", destination)
	}
	if _, err := time.Parse("2006-01-02", dateStr); err != nil {
		return nil, fmt.Errorf("invalid date (expected YYYY-MM-DD): %s", dateStr)
	}

	firstLegs := filterByDepartureDate(s.FlightsByOrigin[origin], dateStr)
	var out []ItineraryOut

	// Direct
	for _, f := range firstLegs {
		if f.Destination == destination {
			it, err := buildItinerary(s, []Flight{f})
			if err != nil {
				return nil, err
			}
			out = append(out, it)
		}
	}

	// 1-stop
	for _, f1 := range firstLegs {
		conn := f1.Destination
		for _, f2 := range s.FlightsByOrigin[conn] {
			if f2.Destination != destination {
				continue
			}
			ok, lay, err := validConnection(s, f1, f2)
			if err != nil {
				return nil, err
			}
			if !ok {
				continue
			}
			it, err := buildItinerary(s, []Flight{f1, f2})
			if err != nil {
				return nil, err
			}
			it.LayoversMinutes = []int{lay}
			out = append(out, it)
		}
	}

	// 2-stop
	for _, f1 := range firstLegs {
		x := f1.Destination
		for _, f2 := range s.FlightsByOrigin[x] {
			ok12, lay1, err := validConnection(s, f1, f2)
			if err != nil {
				return nil, err
			}
			if !ok12 {
				continue
			}
			y := f2.Destination
			for _, f3 := range s.FlightsByOrigin[y] {
				if f3.Destination != destination {
					continue
				}
				ok23, lay2, err := validConnection(s, f2, f3)
				if err != nil {
					return nil, err
				}
				if !ok23 {
					continue
				}
				it, err := buildItinerary(s, []Flight{f1, f2, f3})
				if err != nil {
					return nil, err
				}
				it.LayoversMinutes = []int{lay1, lay2}
				out = append(out, it)
			}
		}
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].TotalDurationMinutes < out[j].TotalDurationMinutes
	})
	return out, nil
}

func filterByDepartureDate(flights []Flight, dateStr string) []Flight {
	var res []Flight
	for _, f := range flights {
		if strings.HasPrefix(f.DepartureTime, dateStr) {
			res = append(res, f)
		}
	}
	return res
}

func buildItinerary(s *Store, flights []Flight) (ItineraryOut, error) {
	seg := make([]SegmentOut, 0, len(flights))
	totalPrice := 0.0

	firstDep, err := DepartureUTC(s, flights[0])
	if err != nil {
		return ItineraryOut{}, err
	}
	lastArr, err := ArrivalUTC(s, flights[len(flights)-1])
	if err != nil {
		return ItineraryOut{}, err
	}

	for _, f := range flights {
		totalPrice += float64(f.Price)
		seg = append(seg, SegmentOut{
			FlightNumber:   f.FlightNumber,
			Airline:        f.Airline,
			Origin:         f.Origin,
			Destination:    f.Destination,
			DepartureLocal: f.DepartureTime,
			ArrivalLocal:   f.ArrivalTime,
			Price:          float64(f.Price),
			Aircraft:       f.Aircraft,
		})
	}

	totalDurMin := Minutes(lastArr.Sub(firstDep))
	if totalDurMin < 0 {
		totalDurMin = 0
	}

	return ItineraryOut{
		Segments:             seg,
		LayoversMinutes:      []int{},
		TotalDurationMinutes: totalDurMin,
		TotalPrice:           round2(totalPrice),
	}, nil
}

func round2(x float64) float64 { return float64(int(x*100+0.5)) / 100 }

func validConnection(s *Store, f1, f2 Flight) (bool, int, error) {
	if f1.Destination != f2.Origin {
		return false, 0, nil
	}

	arr1, err := ArrivalUTC(s, f1)
	if err != nil {
		return false, 0, err
	}
	dep2, err := DepartureUTC(s, f2)
	if err != nil {
		return false, 0, err
	}

	lay := dep2.Sub(arr1)
	layMin := Minutes(lay)
	if layMin < 0 {
		return false, 0, nil
	}
	if lay > 6*time.Hour {
		return false, 0, nil
	}

	minLay := 45 * time.Minute
	if isInternationalConnection(s, f1, f2) {
		minLay = 90 * time.Minute
	}
	if lay < minLay {
		return false, 0, nil
	}

	return true, layMin, nil
}

func isInternationalConnection(s *Store, f1, f2 Flight) bool {
	c1o, ok := s.Country(f1.Origin)
	if !ok {
		return true
	}
	c1d, ok := s.Country(f1.Destination)
	if !ok {
		return true
	}
	c2o, ok := s.Country(f2.Origin)
	if !ok {
		return true
	}
	c2d, ok := s.Country(f2.Destination)
	if !ok {
		return true
	}

	f1Domestic := c1o == c1d
	f2Domestic := c2o == c2d

	if f1Domestic && f2Domestic && c1d == c2o {
		return false
	}
	return true
}
