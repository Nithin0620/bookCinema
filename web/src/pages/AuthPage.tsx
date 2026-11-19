import { useState, type FormEvent } from "react";
import { Link, useLocation, useNavigate } from "react-router-dom";
import { useAuth } from "../AuthContext";
import { errorMessage } from "../api";

export function AuthPage({ mode }: { mode: "login" | "register" }) {
  const { login, register } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);

  const from = (location.state as { from?: string } | null)?.from ?? "/";

  const submit = async (event: FormEvent) => {
    event.preventDefault();
    setBusy(true);
    setError("");
    try {
      if (mode === "login") {
        await login(email, password);
      } else {
        await register(email, password);
      }
      navigate(from, { replace: true });
    } catch (cause: unknown) {
      setError(errorMessage(cause));
    } finally {
      setBusy(false);
    }
  };

  return (
    <section className="auth-page">
      <div className="auth-card">
        <h1>{mode === "login" ? "Welcome back" : "Create an account"}</h1>
        <p className="muted">
          {mode === "login" ? "Log in to hold and confirm seats." : "Sign up to start booking."}
        </p>
        {error && <div className="alert error">{error}</div>}
        <form onSubmit={(event) => void submit(event)}>
          <label>
            Email
            <input
              type="email"
              value={email}
              onChange={(event) => setEmail(event.target.value)}
              required
              autoComplete="email"
            />
          </label>
          <label>
            Password
            <input
              type="password"
              value={password}
              onChange={(event) => setPassword(event.target.value)}
              required
              minLength={8}
              autoComplete={mode === "login" ? "current-password" : "new-password"}
            />
          </label>
          <button type="submit" className="button full" disabled={busy}>
            {busy ? "Please wait..." : mode === "login" ? "Log in" : "Sign up"}
          </button>
        </form>
        <p className="muted small">
          {mode === "login" ? (
            <>
              New here? <Link to="/register">Create an account</Link>
            </>
          ) : (
            <>
              Already have an account? <Link to="/login">Log in</Link>
            </>
          )}
        </p>
      </div>
    </section>
  );
}
