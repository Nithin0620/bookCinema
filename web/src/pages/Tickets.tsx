import { useEffect, useState } from "react";
import { api, errorMessage } from "../api";
import { formatDate, formatPrice } from "../format";
import type { Showtime, TicketDetail } from "../types";

export function Tickets() {
  const [tickets, setTickets] = useState<TicketDetail[]>([]);
  const [showtimes, setShowtimes] = useState<Showtime[]>([]);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let active = true;
    Promise.all([api.myTickets(), api.showtimes()])
      .then(([ticketList, showtimeList]) => {
        if (!active) return;
        setTickets(ticketList);
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

  if (loading) return <div className="state">Loading tickets...</div>;

  return (
    <section>
      <div className="page-heading">
        <h1>Your tickets</h1>
        <p>Show the QR code at the cinema entrance.</p>
      </div>
      {error && <div className="alert error">{error}</div>}
      {tickets.length === 0 ? (
        <div className="state">No tickets yet. Confirm a booking to get one.</div>
      ) : (
        <div className="ticket-grid">
          {tickets.map(({ ticket, booking }) => {
            const showtime = showtimes.find((item) => item.id === booking.showtime_id);
            return (
              <article className="ticket-card" key={ticket.id}>
                <img
                  src={`/api/tickets/${ticket.id}/qr.png`}
                  alt={`QR code for seat ${booking.seat_id}`}
                />
                <div>
                  <h2>Seat {booking.seat_id}</h2>
                  <p className="muted">
                    {showtime ? formatDate(showtime.starts_at) : "Showtime"} ·{" "}
                    {showtime?.screen_name ?? booking.showtime_id}
                  </p>
                  <p className="muted small">Code {ticket.code}</p>
                  <strong>{formatPrice(booking.price_cents)}</strong>
                </div>
              </article>
            );
          })}
        </div>
      )}
    </section>
  );
}
