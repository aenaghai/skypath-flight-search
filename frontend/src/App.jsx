import React, { useMemo, useState } from 'react'

function fmtMinutes(m) {
  const h = Math.floor(m / 60)
  const r = m % 60
  if (h <= 0) return `${r}m`
  if (r === 0) return `${h}h`
  return `${h}h ${r}m`
}

function money(x) {
  return new Intl.NumberFormat(undefined, { style: 'currency', currency: 'USD' }).format(x)
}

function segmentLine(seg) {
  return `${seg.origin} → ${seg.destination} • ${seg.departureLocal} → ${seg.arrivalLocal} • ${seg.flightNumber}`
}

export default function App() {
  const [origin, setOrigin] = useState('JFK')
  const [destination, setDestination] = useState('LAX')
  const [date, setDate] = useState('2024-03-15')

  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [results, setResults] = useState(null)

  const canSearch = useMemo(() => {
    return origin.trim().length === 3 && destination.trim().length === 3 && /^\d{4}-\d{2}-\d{2}$/.test(date)
  }, [origin, destination, date])

  async function onSubmit(e) {
    e.preventDefault()
    setError('')
    setResults(null)

    const o = origin.trim().toUpperCase()
    const d = destination.trim().toUpperCase()

    if (o.length !== 3 || d.length !== 3) {
      setError('Airport codes must be 3 letters (e.g., JFK).')
      return
    }
    if (o === d) {
      setError('Origin and destination cannot be the same.')
      return
    }
    if (!/^\d{4}-\d{2}-\d{2}$/.test(date)) {
      setError('Date must be YYYY-MM-DD.')
      return
    }

    setLoading(true)
    try {
      const res = await fetch(`/api/search?origin=${encodeURIComponent(o)}&destination=${encodeURIComponent(d)}&date=${encodeURIComponent(date)}`)
      const body = await res.json()
      if (!res.ok) {
        setError(body?.message || 'API error')
        return
      }
      setResults(body)
    } catch {
      setError('Failed to reach backend. Is docker-compose running?')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="page">
      <header className="header">
        <h1>SkyPath Flight Search</h1>
        <p className="muted">Direct, 1-stop, and 2-stop itineraries. Sorted by total travel time.</p>
      </header>

      <section className="card">
        <form onSubmit={onSubmit} className="form">
          <label>
            Origin
            <input value={origin} onChange={(e) => setOrigin(e.target.value)} placeholder="JFK" maxLength={3} />
          </label>
          <label>
            Destination
            <input value={destination} onChange={(e) => setDestination(e.target.value)} placeholder="LAX" maxLength={3} />
          </label>
          <label>
            Date
            <input value={date} onChange={(e) => setDate(e.target.value)} placeholder="2024-03-15" />
          </label>

          <button type="submit" disabled={!canSearch || loading}>
            {loading ? 'Searching…' : 'Search'}
          </button>
        </form>

        {error && <div className="error">{error}</div>}
        {!loading && results && results.count === 0 && (
          <div className="empty">No itineraries found for this search.</div>
        )}
      </section>

      {results && results.count > 0 && (
        <section className="results">
          <h2>Results ({results.count})</h2>

          {results.itineraries.map((it, idx) => (
            <div className="card itinerary" key={idx}>
              <div className="it-head">
                <div><strong>Total duration:</strong> {fmtMinutes(it.totalDurationMinutes)}</div>
                <div><strong>Total price:</strong> {money(it.totalPrice)}</div>
                <div><strong>Stops:</strong> {Math.max(0, it.segments.length - 1)}</div>
              </div>

              <ol className="segments">
                {it.segments.map((s, i) => (
                  <li key={i} className="segment">
                    <div className="seg-title">{segmentLine(s)}</div>
                    <div className="muted small">{s.airline} • {s.aircraft} • {money(s.price)}</div>

                    {i < it.layoversMinutes.length && (
                      <div className="layover">
                        Layover at <strong>{s.destination}</strong>: {fmtMinutes(it.layoversMinutes[i])}
                      </div>
                    )}
                  </li>
                ))}
              </ol>
            </div>
          ))}
        </section>
      )}

      <footer className="footer muted">
        Tip: Try JFK → LAX or SFO → NRT for 2024-03-15.
      </footer>
    </div>
  )
}
