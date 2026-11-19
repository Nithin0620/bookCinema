import { useEffect, useState } from "react";
import { api, errorMessage } from "../api";
import { formatDate } from "../format";
import type { AppNotification } from "../types";

export function Notifications() {
  const [items, setItems] = useState<AppNotification[]>([]);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    let active = true;
    api
      .notifications()
      .then((result) => {
        if (active) setItems(result);
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

  const markAllRead = async () => {
    setBusy(true);
    try {
      await api.markNotificationsRead();
      setItems((current) => current.map((item) => ({ ...item, read_at: new Date().toISOString() })));
      setError("");
    } catch (cause: unknown) {
      setError(errorMessage(cause));
    } finally {
      setBusy(false);
    }
  };

  if (loading) return <div className="state">Loading notifications...</div>;

  return (
    <section>
      <div className="page-heading page-heading-row">
        <div>
          <h1>Notifications</h1>
          <p>Booking confirmations and cancellations appear here.</p>
        </div>
        <button
          type="button"
          className="button small"
          disabled={busy || items.every((item) => item.read_at !== null)}
          onClick={() => void markAllRead()}
        >
          {busy ? "Updating..." : "Mark all read"}
        </button>
      </div>
      {error && <div className="alert error">{error}</div>}
      {items.length === 0 ? (
        <div className="state">No notifications yet.</div>
      ) : (
        <div className="booking-list">
          {items.map((item) => (
            <article className={`notification-card ${item.read_at ? "" : "unread"}`} key={item.id}>
              <div>
                <h2>{item.title}</h2>
                <p>{item.body}</p>
              </div>
              <small className="muted">{formatDate(item.created_at)}</small>
            </article>
          ))}
        </div>
      )}
    </section>
  );
}
