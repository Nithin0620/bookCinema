import { useEffect, useState } from "react";
import { api, errorMessage } from "../api";
import type { Movie, Showtime } from "../types";

export function Admin() {
  const [movies, setMovies] = useState<Movie[]>([]);
  const [showtimes, setShowtimes] = useState<Showtime[]>([]);
  const [error, setError] = useState("");
  const [message, setMessage] = useState("");
  const [busy, setBusy] = useState("");

  const [movie, setMovie] = useState({
    title: "",
    description: "",
    duration_min: "120",
    poster_url: "",
  });
  const [showtime, setShowtime] = useState({
    movie_id: "",
    starts_at: "",
    screen_name: "Screen 1",
    rows: "8",
    seats_per_row: "12",
    premium_rows: "3",
    premium_price_cents: "1500",
    standard_price_cents: "1000",
  });

  useEffect(() => {
    let active = true;
    Promise.all([api.movies(), api.showtimes()])
      .then(([movieList, showtimeList]) => {
        if (!active) return;
        setMovies(movieList);
        setShowtimes(showtimeList);
        setShowtime((current) => ({
          ...current,
          movie_id: current.movie_id || movieList[0]?.id || "",
        }));
      })
      .catch((cause: unknown) => {
        if (active) setError(errorMessage(cause));
      });
    return () => {
      active = false;
    };
  }, []);

  const createMovie = async () => {
    setBusy("movie");
    setError("");
    setMessage("");
    try {
      const created = await api.createMovie({
        title: movie.title,
        description: movie.description,
        duration_min: Number(movie.duration_min),
        poster_url: movie.poster_url,
      });
      setMovies((current) => [created, ...current]);
      setShowtime((current) => ({ ...current, movie_id: current.movie_id || created.id }));
      setMovie({ title: "", description: "", duration_min: "120", poster_url: "" });
      setMessage(`Movie "${created.title}" created.`);
    } catch (cause: unknown) {
      setError(errorMessage(cause));
    } finally {
      setBusy("");
    }
  };

  const createShowtime = async () => {
    setBusy("showtime");
    setError("");
    setMessage("");
    try {
      const created = await api.createShowtime({
        movie_id: showtime.movie_id,
        starts_at: new Date(showtime.starts_at).toISOString(),
        screen_name: showtime.screen_name,
        rows: Number(showtime.rows),
        seats_per_row: Number(showtime.seats_per_row),
        premium_rows: Number(showtime.premium_rows),
        premium_price_cents: Number(showtime.premium_price_cents),
        standard_price_cents: Number(showtime.standard_price_cents),
      });
      setShowtimes((current) => [created, ...current]);
      setShowtime((current) => ({ ...current, starts_at: "" }));
      setMessage("Showtime created.");
    } catch (cause: unknown) {
      setError(errorMessage(cause));
    } finally {
      setBusy("");
    }
  };

  return (
    <section>
      <div className="page-heading">
        <h1>Admin</h1>
        <p>Add movies and schedule showtimes.</p>
      </div>
      {error && <div className="alert error">{error}</div>}
      {message && <div className="alert success">{message}</div>}

      <div className="admin-grid">
        <article className="panel">
          <h2>New movie</h2>
          <label>
            Title
            <input
              value={movie.title}
              onChange={(event) => setMovie({ ...movie, title: event.target.value })}
            />
          </label>
          <label>
            Description
            <textarea
              value={movie.description}
              onChange={(event) => setMovie({ ...movie, description: event.target.value })}
            />
          </label>
          <label>
            Duration (minutes)
            <input
              type="number"
              min="1"
              value={movie.duration_min}
              onChange={(event) => setMovie({ ...movie, duration_min: event.target.value })}
            />
          </label>
          <label>
            Poster URL
            <input
              value={movie.poster_url}
              onChange={(event) => setMovie({ ...movie, poster_url: event.target.value })}
            />
          </label>
          <button
            type="button"
            className="button full"
            disabled={busy === "movie" || movie.title.trim() === ""}
            onClick={() => void createMovie()}
          >
            {busy === "movie" ? "Creating..." : "Create movie"}
          </button>
        </article>

        <article className="panel">
          <h2>New showtime</h2>
          <label>
            Movie
            <select
              value={showtime.movie_id}
              onChange={(event) => setShowtime({ ...showtime, movie_id: event.target.value })}
            >
              {movies.map((item) => (
                <option value={item.id} key={item.id}>
                  {item.title}
                </option>
              ))}
            </select>
          </label>
          <label>
            Starts at
            <input
              type="datetime-local"
              value={showtime.starts_at}
              onChange={(event) => setShowtime({ ...showtime, starts_at: event.target.value })}
            />
          </label>
          <label>
            Screen name
            <input
              value={showtime.screen_name}
              onChange={(event) => setShowtime({ ...showtime, screen_name: event.target.value })}
            />
          </label>
          <div className="form-row">
            <label>
              Rows
              <input
                type="number"
                min="1"
                value={showtime.rows}
                onChange={(event) => setShowtime({ ...showtime, rows: event.target.value })}
              />
            </label>
            <label>
              Seats per row
              <input
                type="number"
                min="1"
                value={showtime.seats_per_row}
                onChange={(event) => setShowtime({ ...showtime, seats_per_row: event.target.value })}
              />
            </label>
          </div>
          <div className="form-row">
            <label>
              Premium rows
              <input
                type="number"
                min="0"
                value={showtime.premium_rows}
                onChange={(event) => setShowtime({ ...showtime, premium_rows: event.target.value })}
              />
            </label>
            <label>
              Premium price (cents)
              <input
                type="number"
                min="0"
                value={showtime.premium_price_cents}
                onChange={(event) =>
                  setShowtime({ ...showtime, premium_price_cents: event.target.value })
                }
              />
            </label>
          </div>
          <label>
            Standard price (cents)
            <input
              type="number"
              min="0"
              value={showtime.standard_price_cents}
              onChange={(event) =>
                setShowtime({ ...showtime, standard_price_cents: event.target.value })
              }
            />
          </label>
          <button
            type="button"
            className="button full"
            disabled={
              busy === "showtime" || showtime.movie_id === "" || showtime.starts_at === ""
            }
            onClick={() => void createShowtime()}
          >
            {busy === "showtime" ? "Creating..." : "Create showtime"}
          </button>
        </article>
      </div>

      <article className="panel">
        <h2>Scheduled showtimes ({showtimes.length})</h2>
        <div className="booking-list">
          {showtimes.slice(0, 10).map((item) => (
            <div className="booking-card" key={item.id}>
              <div>
                <strong>{item.screen_name}</strong>
                <small className="muted">{new Date(item.starts_at).toLocaleString()}</small>
              </div>
              <small className="muted">
                {item.rows} × {item.seats_per_row} seats · {item.premium_rows} premium rows
              </small>
            </div>
          ))}
        </div>
      </article>
    </section>
  );
}
