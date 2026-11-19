import { useEffect, useState, type ReactNode } from "react";
import { Link, useLocation, useNavigate } from "react-router-dom";
import { api } from "../api";
import { useAuth } from "../AuthContext";
import { DotField } from "./DotField";

export function Layout({ children }: { children: ReactNode }) {
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const [unread, setUnread] = useState(0);

  useEffect(() => {
    if (!user) {
      setUnread(0);
      return;
    }
    let active = true;
    const load = () => {
      api
        .notifications()
        .then((items) => {
          if (active) {
            setUnread(items.filter((item) => item.read_at === null).length);
          }
        })
        .catch(() => undefined);
    };
    load();
    const timer = setInterval(load, 5000);
    return () => {
      active = false;
      clearInterval(timer);
    };
  }, [user, location.pathname]);

  const onLogout = async () => {
    await logout();
    navigate("/");
  };

  return (
    <div className="app-shell">
      <DotField />
      <header className="topbar">
        <Link to="/" className="brand">
          book<span>Cinema</span>
        </Link>
        <nav>
          <Link to="/">Movies</Link>
          {user && (
            <>
              <Link to="/bookings">My bookings</Link>
              <Link to="/tickets">Tickets</Link>
              <Link to="/notifications">
                Notifications{unread > 0 ? ` (${unread})` : ""}
              </Link>
              {user.is_admin && <Link to="/admin">Admin</Link>}
              <span className="user-email">{user.email}</span>
              <button type="button" className="link-button" onClick={onLogout}>
                Log out
              </button>
            </>
          )}
          {!user && (
            <>
              <Link to="/login">Log in</Link>
              <Link to="/register" className="button small">
                Sign up
              </Link>
            </>
          )}
        </nav>
      </header>
      <main className="main-content">{children}</main>
      <footer className="footer">
        <div>
          <Link to="/" className="footer-brand" aria-label="bookCinema">
            <span className="mark">
              <svg viewBox="0 0 20 20" aria-hidden="true">
                <circle cx="6" cy="6" r="3" />
                <circle cx="14" cy="6" r="3" />
                <circle cx="6" cy="14" r="3" />
                <circle cx="14" cy="14" r="3" />
              </svg>
            </span>
            bookCinema
          </Link>
          <p className="tag">
            A calmer way to book: two-minute holds, tickets and reminders in one
            place.
          </p>
          <nav className="social" aria-label="Social">
            <a href="#" aria-label="LinkedIn">
              <svg viewBox="0 0 24 24">
                <rect x="3" y="3" width="18" height="18" rx="3" />
                <path d="M8 10v6M8 7.5v.01M12 16v-3.5a2 2 0 0 1 4 0V16M12 10v6" />
              </svg>
            </a>
            <a href="#" aria-label="X">
              <svg viewBox="0 0 24 24">
                <path d="M4 4l16 16M20 4L4 20" />
              </svg>
            </a>
            <a href="#" aria-label="Instagram">
              <svg viewBox="0 0 24 24">
                <rect x="3" y="3" width="18" height="18" rx="5" />
                <circle cx="12" cy="12" r="4" />
                <path d="M17 7v.01" />
              </svg>
            </a>
            <a href="#" aria-label="YouTube">
              <svg viewBox="0 0 24 24">
                <rect x="3" y="5" width="18" height="14" rx="4" />
                <path d="M10 9.5v5l4.5-2.5Z" />
              </svg>
            </a>
          </nav>
          <p className="copy">© bookCinema 2026 · All rights reserved</p>
        </div>
        <div className="links">
          <div>
            <h3>Explore</h3>
            <ul>
              <li><Link to="/">Movies</Link></li>
              <li><Link to="/tickets">Tickets</Link></li>
              <li><Link to="/notifications">Notifications</Link></li>
            </ul>
          </div>
          <div>
            <h3>Account</h3>
            <ul>
              <li><Link to="/bookings">My bookings</Link></li>
              <li><Link to="/login">Log in</Link></li>
              <li><Link to="/register">Sign up</Link></li>
              {user?.is_admin && <li><Link to="/admin">Admin</Link></li>}
            </ul>
          </div>
          <div>
            <h3>System</h3>
            <ul>
              <li><span>Redis holds</span></li>
              <li><span>Postgres bookings</span></li>
              <li><span>Go + React</span></li>
            </ul>
          </div>
        </div>
      </footer>
    </div>
  );
}
