package service

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"skypath/backend/models"
	"skypath/backend/repository"
	"skypath/backend/utils"
)

type SearchService struct {
	Store *repository.Store
}

func NewSearchService(store *repository.Store) *SearchService {
	return &SearchService{Store: store}
}

func (ss *SearchService) SearchItineraries(origin, destination, dateStr string) ([]models.ItineraryOut, error) {
	origin = utils.UpperCompact(origin)
	destination = utils.UpperCompact(destination)

	// PDF test case allows empty results OR validation error
	if origin == destination {
		return []models.ItineraryOut{}, nil
	}

	if !utils.IsIATACode(origin) {
		return nil, fmt.Errorf("invalid origin airport code: %s", origin)
	}
	if !utils.IsIATACode(destination) {
		return nil, fmt.Errorf("invalid destination airport code: %s", destination)
	}

	if _, ok := ss.Store.GetAirport(origin); !ok {
		return nil, fmt.Errorf("invalid origin airport code: %s", origin)
	}
	if _, ok := ss.Store.GetAirport(destination); !ok {
		return nil, fmt.Errorf("invalid destination airport code: %s", destination)
	}

	if _, err := time.Parse("2006-01-02", dateStr); err != nil {
		return nil, fmt.Errorf("invalid date (expected YYYY-MM-DD): %s", dateStr)
	}

	firstLegs := filterByDepartureDate(ss.Store.FlightsByOrigin[origin], dateStr)
	var out []models.ItineraryOut

	// Direct
	for _, f := range firstLegs {
		if f.Destination == destination {
			it, err := buildItinerary(ss.Store, []models.Flight{f})
			if err != nil {
				return nil, err
			}
			out = append(out, it)
		}
	}

	// 1-stop
	for _, f1 := range firstLegs {
		conn := f1.Destination
		for _, f2 := range ss.Store.FlightsByOrigin[conn] {
			if f2.Destination != destination {
				continue
			}
			ok, lay, err := validConnection(ss.Store, f1, f2)
			if err != nil {
				return nil, err
			}
			if !ok {
				continue
			}
			it, err := buildItinerary(ss.Store, []models.Flight{f1, f2})
			if err != nil {
				return nil, err
			}
			it.LayoversMinutes = []int{lay}
			out = append(out, it)
		}
	}

	// 2-stop max
	for _, f1 := range firstLegs {
		x := f1.Destination
		for _, f2 := range ss.Store.FlightsByOrigin[x] {
			ok12, lay1, err := validConnection(ss.Store, f1, f2)
			if err != nil {
				return nil, err
			}
			if !ok12 {
				continue
			}
			y := f2.Destination
			for _, f3 := range ss.Store.FlightsByOrigin[y] {
				if f3.Destination != destination {
					continue
				}
				ok23, lay2, err := validConnection(ss.Store, f2, f3)
				if err != nil {
					return nil, err
				}
				if !ok23 {
					continue
				}
				it, err := buildItinerary(ss.Store, []models.Flight{f1, f2, f3})
				if err != nil {
					return nil, err
				}
				it.LayoversMinutes = []int{lay1, lay2}
				out = append(out, it)
			}
		}
	}

	// PDF: sort by total travel time shortest first
	sort.Slice(out, func(i, j int) bool {
		return out[i].TotalDurationMinutes < out[j].TotalDurationMinutes
	})

	return out, nil
}

func filterByDepartureDate(flights []models.Flight, dateStr string) []models.Flight {
	var res []models.Flight
	for _, f := range flights {
		if strings.HasPrefix(f.DepartureTime, dateStr) {
			res = append(res, f)
		}
	}
	return res
}

func buildItinerary(s *repository.Store, flights []models.Flight) (models.ItineraryOut, error) {
	seg := make([]models.SegmentOut, 0, len(flights))
	totalPrice := 0.0

	firstDep, err := departureUTC(s, flights[0])
	if err != nil {
		return models.ItineraryOut{}, err
	}
	lastArr, err := arrivalUTC(s, flights[len(flights)-1])
	if err != nil {
		return models.ItineraryOut{}, err
	}

	for _, f := range flights {
		totalPrice += float64(f.Price)
		seg = append(seg, models.SegmentOut{
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

	totalDurMin := utils.Minutes(lastArr.Sub(firstDep))
	if totalDurMin < 0 {
		totalDurMin = 0
	}

	return models.ItineraryOut{
		Segments:             seg,
		LayoversMinutes:      []int{},
		TotalDurationMinutes: totalDurMin,
		TotalPrice:           round2(totalPrice),
	}, nil
}

func round2(x float64) float64 { return float64(int(x*100+0.5)) / 100 }

func departureUTC(s *repository.Store, f models.Flight) (time.Time, error) {
	loc, err := s.TZLocation(f.Origin)
	if err != nil {
		return time.Time{}, err
	}
	t, err := utils.ParseLocalAt(f.DepartureTime, loc)
	if err != nil {
		return time.Time{}, err
	}
	return t.UTC(), nil
}

func arrivalUTC(s *repository.Store, f models.Flight) (time.Time, error) {
	loc, err := s.TZLocation(f.Destination)
	if err != nil {
		return time.Time{}, err
	}
	t, err := utils.ParseLocalAt(f.ArrivalTime, loc)
	if err != nil {
		return time.Time{}, err
	}
	return t.UTC(), nil
}

func validConnection(s *repository.Store, f1, f2 models.Flight) (bool, int, error) {
	// PDF: cannot change airports during layover
	if f1.Destination != f2.Origin {
		return false, 0, nil
	}

	arr1, err := arrivalUTC(s, f1)
	if err != nil {
		return false, 0, err
	}
	dep2, err := departureUTC(s, f2)
	if err != nil {
		return false, 0, err
	}

	lay := dep2.Sub(arr1)
	layMin := utils.Minutes(lay)
	if layMin < 0 {
		return false, 0, nil
	}

	// PDF: max layover 6 hours
	if lay > 6*time.Hour {
		return false, 0, nil
	}

	// PDF: min layover 45 (domestic) / 90 (international)
	minLay := 45 * time.Minute
	if isInternationalConnection(s, f1, f2) {
		minLay = 90 * time.Minute
	}
	if lay < minLay {
		return false, 0, nil
	}

	return true, layMin, nil
}

// PDF definition: connection is "domestic" if BOTH arriving and departing flights are within the same country
func isInternationalConnection(s *repository.Store, f1, f2 models.Flight) bool {
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

	// Domestic connection only if both flights are domestic AND same country
	if f1Domestic && f2Domestic && c1d == c2o {
		return false
	}
	return true
}
