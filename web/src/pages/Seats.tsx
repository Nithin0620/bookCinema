import { useCallback, useEffect, useMemo, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { api, errorMessage } from "../api";
import { formatDate, formatPrice, formatRemaining } from "../format";
import type { Challenge, SeatView, SeatsResponse } from "../types";

export function Seats() {
  const { id } = useParams();
  const showtimeId = id ?? "";
  const navigate = useNavigate();
  const [data, setData] = useState<SeatsResponse | null>(null);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState("");
  const [checkoutBusy, setCheckoutBusy] = useState(false);
  const [challenge, setChallenge] = useState<Challenge | null>(null);
  const [answer, setAnswer] = useState("");
  const [, setTick] = useState(0);

  const load = useCallback(async () => {
    const result = await api.seats(showtimeId);
    const expired = result.seats.filter(
      (seat) =>
        seat.status === "held" &&
        seat.mine &&
        seat.session_id &&
        seat.expires_at &&
        new Date(seat.expires_at).getTime() <= Date.now(),
    );
    if (expired.length > 0) {
      await Promise.all(
        expired.map((seat) => api.release(seat.session_id as string)),
      );
      setData(await api.seats(showtimeId));
      return;
    }
    setData(result);
  }, [showtimeId]);

  useEffect(() => {
    let active = true;
    const poll = () => {
      load()
        .then(() => {
          if (active) setError("");
        })
        .catch((cause: unknown) => {
          if (active) setError(errorMessage(cause));
        });
    };

    poll();
    const seatPoll = setInterval(poll, 2000);
    const clock = setInterval(() => setTick((value) => value + 1), 1000);
    return () => {
      active = false;
      clearInterval(seatPoll);
      clearInterval(clock);
    };
  }, [load]);

  const run = useCallback(
    async (key: string, action: () => Promise<void>) => {
      setBusy(key);
      try {
        await action();
        await load();
        setError("");
      } catch (cause: unknown) {
        setError(errorMessage(cause));
      } finally {
        setBusy("");
      }
    },
    [load],
  );

  const onSeatClick = (seat: SeatView) => {
    if (busy) return;
    if (seat.status === "available") {
      void run(seat.seat_id, () =>
        api.hold(showtimeId, seat.seat_id).then(() => undefined),
      );
    } else if (seat.status === "held" && seat.mine && seat.session_id) {
      void run(seat.seat_id, () => api.release(seat.session_id as string));
    }
  };

  const cart = useMemo(
    () =>
      (data?.seats ?? []).filter(
        (seat) => seat.status === "held" && seat.mine && seat.session_id,
      ),
    [data],
  );

  const rows = useMemo(() => {
    const grouped = new Map<string, SeatView[]>();
    for (const seat of data?.seats ?? []) {
      const row = seat.seat_id[0];
      grouped.set(row, [...(grouped.get(row) ?? []), seat]);
    }
    return [...grouped.entries()];
  }, [data]);

  const resetCheckout = useCallback(() => {
    setChallenge(null);
    setAnswer("");
    setCheckoutBusy(false);
  }, []);

  const startCheckout = async () => {
    if (cart.length === 0 || checkoutBusy) return;
    setError("");
    try {
      setChallenge(await api.newChallenge());
    } catch (cause: unknown) {
      setError(errorMessage(cause));
    }
  };

  const submitCheckout = async () => {
    if (!challenge || answer.trim() === "") return;
    setCheckoutBusy(true);
    try {
      await api.checkout(
        cart.map((seat) => seat.session_id as string),
        challenge.id,
        answer,
      );
      navigate("/tickets");
    } catch (cause: unknown) {
      setError(errorMessage(cause));
      resetCheckout();
      void startCheckout();
    }
  };

  useEffect(() => {
    if (cart.length === 0 && challenge) {
      resetCheckout();
    }
  }, [cart.length, challenge, resetCheckout]);

  if (!data) {
    return (
      <div className="state">
        {error ? (
          <span className="error">Could not load seats: {error}</span>
        ) : (
          "Loading seats..."
        )}
      </div>
    );
  }

  const total = cart.reduce((sum, seat) => sum + seat.price_cents, 0);
  let previousSection = "";

  return (
    <section>
      <div className="page-heading">
        <h1>Choose your seats</h1>
        <p>
          {formatDate(data.showtime.starts_at)} · {data.showtime.screen_name}
        </p>
      </div>

      {error && <div className="alert error">{error}</div>}

      <div className="booking-layout">
        <div className="screen-panel">
          <div className="screen-label">SCREEN</div>
          <div className="seat-rows">
            {rows.map(([row, seats]) => {
              const section = seats[0]?.section ?? "standard";
              const showLabel = section !== previousSection;
              previousSection = section;
              return (
                <div key={row}>
                  {showLabel && (
                    <div className={`section-label ${section}`}>
                      {section === "premium" ? "Premium" : "Standard"} ·{" "}
                      {formatPrice(
                        section === "premium"
                          ? data.showtime.premium_price_cents
                          : data.showtime.standard_price_cents,
                      )}
                    </div>
                  )}
                  <div className="seat-row">
                    <span className="row-label">{row}</span>
                    <div className="seat-group">
                      {seats.map((seat) => {
                        const className = [
                          "seat",
                          seat.section,
                          seat.status === "confirmed" ? "confirmed" : "",
                          seat.status === "held" && seat.mine ? "mine" : "",
                          seat.status === "held" && !seat.mine ? "other-held" : "",
                          busy === seat.seat_id ? "busy" : "",
                        ]
                          .filter(Boolean)
                          .join(" ");
                        return (
                          <button
                            type="button"
                            className={className}
                            key={seat.seat_id}
                            disabled={
                              seat.status === "confirmed" ||
                              (seat.status === "held" && !seat.mine) ||
                              Boolean(busy)
                            }
                            onClick={() => onSeatClick(seat)}
                            title={`Seat ${seat.seat_id} · ${seat.section} · ${formatPrice(seat.price_cents)}`}
                          >
                            {seat.seat_id}
                          </button>
                        );
                      })}
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
          <div className="legend">
            <span><i className="dot premium" /> Premium</span>
            <span><i className="dot standard" /> Standard</span>
            <span><i className="dot mine" /> Your hold</span>
            <span><i className="dot other-held" /> Other hold</span>
            <span><i className="dot confirmed" /> Confirmed</span>
          </div>
        </div>

        <aside className="cart-panel">
          <h2>Your cart</h2>
          {cart.length === 0 ? (
            <p className="muted">Tap an available seat to place a 2 minute hold.</p>
          ) : (
            <ul className="cart-list">
              {cart.map((seat) => (
                <li key={seat.session_id}>
                  <div>
                    <strong>
                      Seat {seat.seat_id} <small className="muted">{seat.section}</small>
                    </strong>
                    <small className="muted">
                      {formatPrice(seat.price_cents)} · Hold {formatRemaining(seat.expires_at)}
                    </small>
                  </div>
                  <button
                    type="button"
                    className="link-button danger"
                    disabled={Boolean(busy)}
                    onClick={() => onSeatClick(seat)}
                  >
                    Release
                  </button>
                </li>
              ))}
            </ul>
          )}
          <div className="cart-total">
            <span>Total</span>
            <strong>{formatPrice(total)}</strong>
          </div>

          {challenge ? (
            <div className="puzzle-box">
              <p className="small muted">Payment confirmation puzzle</p>
              <strong>{challenge.question}</strong>
              <input
                type="number"
                value={answer}
                onChange={(event) => setAnswer(event.target.value)}
                placeholder="Your answer"
                disabled={checkoutBusy}
              />
              <div className="puzzle-actions">
                <button
                  type="button"
                  className="button full"
                  disabled={checkoutBusy || answer.trim() === ""}
                  onClick={() => void submitCheckout()}
                >
                  {checkoutBusy ? "Confirming..." : "Confirm booking"}
                </button>
                <button
                  type="button"
                  className="link-button"
                  disabled={checkoutBusy}
                  onClick={() => {
                    resetCheckout();
                    void startCheckout();
                  }}
                >
                  New puzzle
                </button>
              </div>
            </div>
          ) : (
            <button
              type="button"
              className="button full"
              disabled={cart.length === 0}
              onClick={() => void startCheckout()}
            >
              Continue to confirmation
            </button>
          )}
          <p className="muted small">
            Seats update every 2 seconds. Expired holds are released automatically.
          </p>
        </aside>
      </div>
    </section>
  );
}
