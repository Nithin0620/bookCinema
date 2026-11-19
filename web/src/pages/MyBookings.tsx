import { useEffect, useState } from "react";
import { api, errorMessage } from "../api";
import { formatDate, formatPrice } from "../format";
import type { Booking, Showtime } from "../types";

export function MyBookings() {
  const [bookings, setBookings] = useState<Booking[]>([]);
  const [showtimes, setShowtimes] = useState<Showtime[]>([]);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState("");

  useEffect(() => {
    let active = true;
    Promise.all([api.myBookings(), api.showtimes()])
      .then(([bookingList, showtimeList]) => {
        if (!active) return;
        setBookings(bookingList);
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

  const cancel = async (bookingId: string) => {
    setBusy(bookingId);
    try {
      await api.cancelBooking(bookingId);
      setBookings((current) => current.filter((booking) => booking.id !== bookingId));
      setError("");
    } catch (cause: unknown) {
      setError(errorMessage(cause));
    } finally {
      setBusy("");
    }
  };

  if (loading) return <div className="state">Loading bookings...</div>;

  return (
    <section>
      <div className="page-heading">
        <h1>My bookings</h1>
        <p>Your confirmed seats are stored in Postgres.</p>
      </div>
      {error && <div className="alert error">{error}</div>}
      {bookings.length === 0 ? (
        <div className="state">No bookings yet. Pick a movie to get started.</div>
      ) : (
        <div className="booking-list">
          {bookings.map((booking) => {
            const showtime = showtimes.find((item) => item.id === booking.showtime_id);
            return (
              <article className="booking-card" key={booking.id}>
                <div>
                  <h2>Seat {booking.seat_id}</h2>
                  <p className="muted">
                    {showtime ? formatDate(showtime.starts_at) : "Showtime"} ·{" "}
                    {showtime?.screen_name ?? booking.showtime_id}
                  </p>
                </div>
                <div className="booking-card-actions">
                  <strong>{formatPrice(booking.price_cents)}</strong>
                  <button
                    type="button"
                    className="button small danger-outline"
                    disabled={Boolean(busy)}
                    onClick={() => void cancel(booking.id)}
                  >
                    {busy === booking.id ? "Cancelling..." : "Cancel"}
                  </button>
                </div>
              </article>
            );
          })}
        </div>
      )}
    </section>
  );
}
