# SkyPath – Flight Search Engine
SkyPath is a flight search application that computes valid flight itineraries (direct, 1-stop, and 2-stop) between two airports on a given date, following real-world connection and layover constraints.

The project consists of:
- **Backend**: Go (REST API, in-memory dataset, itinerary computation)
- **Frontend**: React (Vite) served via Nginx
- **Infrastructure**: Docker + docker-compose

---

## Features

### Backend
- Loads flight and airport data from a static `flights.json` dataset at startup
- Exposes a REST API to search flights by:
  - origin airport (IATA code)
  - destination airport (IATA code)
  - date (`YYYY-MM-DD`)
- Computes valid itineraries with:
  - Direct flights
  - One-stop connections
  - Two-stop connections (maximum)
- Enforces realistic connection rules:
  - No airport change during layovers
  - Minimum layover:
    - **45 minutes** for domestic connections
    - **90 minutes** for international connections
  - Maximum layover: **6 hours**
- Timezone-aware calculations (local times → UTC)
- Results sorted by **total travel duration (shortest first)**

### Frontend
- Search form with origin, destination, and date
- Displays:
  - Flight segments
  - Layover durations
  - Total travel time
  - Total price
- Handles:
  - Loading states
  - Empty results
  - API and validation errors
- Communicates with backend via `/api/*` endpoints

---

## Project Structure

```

skypath-flight-search/
│
├── backend/
│   ├── controller/        # HTTP handlers and routing
│   ├── service/           # Business logic (itinerary computation)
│   ├── repository/        # Dataset loading and lookup
│   ├── models/            # Data models and API response structs
│   ├── utils/             # Time, string, HTTP helpers
│   ├── main.go            # Application entrypoint
│   ├── go.mod
│   └── Dockerfile
│
├── frontend/
│   ├── src/               # React source code
│   ├── nginx.conf         # Nginx reverse proxy configuration
│   ├── Dockerfile
│   └── package.json
│
├── flights.json            # Dataset (mounted into backend container)
├── docker-compose.yml
└── README.md

```

---

## API Specification

### Health Check
```

GET /api/health

````

Response:
```json
{ "ok": true }
````

---

### Search Flights

```
GET /api/search?origin=JFK&destination=LAX&date=2024-03-15
```

#### Query Parameters

| Name        | Description                      |
| ----------- | -------------------------------- |
| origin      | IATA airport code (e.g. JFK)     |
| destination | IATA airport code (e.g. LAX)     |
| date        | Travel date in YYYY-MM-DD format |

#### Successful Response

```json
{
  "origin": "JFK",
  "destination": "LAX",
  "date": "2024-03-15",
  "count": 2,
  "itineraries": [
    {
      "segments": [
        {
          "flightNumber": "AA101",
          "airline": "American Airlines",
          "origin": "JFK",
          "destination": "LAX",
          "departureLocal": "2024-03-15T08:00:00",
          "arrivalLocal": "2024-03-15T11:00:00",
          "price": 320,
          "aircraft": "A320"
        }
      ],
      "layoversMinutes": [],
      "totalDurationMinutes": 360,
      "totalPrice": 320
    }
  ]
}
```

#### Error Response

```json
{
  "message": "invalid origin airport code: XXX"
}
```

---

## Business Rules Implemented

* Maximum of **2 stops** per itinerary
* Layovers:

  * Minimum **45 minutes** for domestic connections
  * Minimum **90 minutes** for international connections
  * Maximum **6 hours**
* Connections must occur at the **same airport**
* Domestic vs international determination is based on airport country
* Time comparisons are done in **UTC** after converting from local airport timezones
* Results are sorted by **total journey duration**

---

## Running the Application

### Prerequisites

* Docker
* Docker Compose

### Start the Application

From the repository root:

```bash
docker-compose up --build
```

Services:

* Backend API: `http://localhost:8080`
* Frontend UI: `http://localhost:3000`

---

## Environment Configuration

The backend uses the following environment variables:

| Variable          | Description                             |
| ----------------- | --------------------------------------- |
| PORT              | HTTP port (default: 8080)               |
| FLIGHTS_DATA_PATH | Path to `flights.json` inside container |

In Docker, the dataset is mounted as:

```
./flights.json → /app/data/flights.json
```

---

## Design Decisions & Tradeoffs

### Why Go for Backend

* Strong concurrency model
* Fast startup and low memory overhead

### In-Memory Dataset

* Simplifies the system for the given constraints
* Dataset is loaded once at startup for fast reads

### No Database Layer

* Repository layer abstracts data access
* Can be easily replaced with a persistent store in the future

### Nginx for Frontend

* Serves static assets efficiently
* Acts as a reverse proxy for backend API calls

---

## Possible Improvements

* Pagination or limiting number of itineraries returned
* Price-based or multi-criteria sorting (price + duration)
* Caching frequent searches
* Support multi-day journeys
* Deploy using Kubernetes 

---
