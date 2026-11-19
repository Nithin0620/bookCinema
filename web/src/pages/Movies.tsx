import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { api, errorMessage } from "../api";
import { Faq } from "../components/Faq";
import { formatDate, formatPrice } from "../format";
import type { Movie, Showtime } from "../types";

export function Movies() {
  const [movies, setMovies] = useState<Movie[]>([]);
  const [showtimes, setShowtimes] = useState<Showtime[]>([]);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let active = true;
    Promise.all([api.movies(), api.showtimes()])
      .then(([movieList, showtimeList]) => {
        if (!active) return;
        setMovies(movieList);
        setShowtimes(showtimeList);
      })
      .catch((cause: unknown) => {
        if (active) setError(errorMessage(cause));
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, []);

  if (loading) return <div className="state">Loading movies...</div>;
  if (error) return <div className="state error">Could not load movies: {error}</div>;

  return (
    <section>
      <div className="page-heading">
        <h1>Now showing</h1>
        <p>Pick a showtime, choose your seats, and check out before the hold expires.</p>
      </div>
      <div className="movie-grid">
        {movies.map((movie) => {
          const times = showtimes.filter((showtime) => showtime.movie_id === movie.id);
          return (
            <article className="movie-card" key={movie.id}>
              <div className="movie-poster" aria-hidden="true">
                {movie.title.slice(0, 1)}
              </div>
              <div className="movie-body">
                <h2>{movie.title}</h2>
                <p className="muted">{movie.description}</p>
                <p className="muted small">
                  {movie.duration_min} min · {times.length} showtime
                  {times.length === 1 ? "" : "s"}
                </p>
                <div className="showtime-list">
                  {times.map((showtime) => (
                    <Link
                      className="showtime-chip"
                      to={`/showtimes/${showtime.id}`}
                      key={showtime.id}
                    >
                      <span>{formatDate(showtime.starts_at)}</span>
                      <span className="muted small">{showtime.screen_name}</span>
                      <strong>
                        {formatPrice(showtime.standard_price_cents)}–{formatPrice(showtime.premium_price_cents)}
                      </strong>
                    </Link>
                  ))}
                </div>
              </div>
            </article>
          );
        })}
      </div>
      <Faq />
    </section>
  );
}
