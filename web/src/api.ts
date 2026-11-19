import type {
  AppNotification,
  Booking,
  Challenge,
  Hold,
  Movie,
  SeatsResponse,
  Showtime,
  TicketDetail,
  User,
} from "./types";

export class ApiError extends Error {
  constructor(
    message: string,
    public status: number,
    public code: string,
  ) {
    super(message);
    this.name = "ApiError";
  }
}

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const response = await fetch(`/api${path}`, {
    ...init,
    credentials: "same-origin",
    headers: {
      "Content-Type": "application/json",
      ...(init.headers ?? {}),
    },
  });

  if (!response.ok) {
    let body: { error?: string } = {};
    try {
      body = await response.json();
    } catch {
      body = {};
    }
    const code = body.error ?? "request_failed";
    throw new ApiError(code, response.status, code);
  }

  if (response.status === 204) {
    return undefined as T;
  }
  return response.json() as Promise<T>;
}

function jsonBody(value: unknown): string {
  return JSON.stringify(value);
}

export const api = {
  me: () => request<User>("/auth/me"),

  register: (email: string, password: string) =>
    request<User>("/auth/register", {
      method: "POST",
      body: jsonBody({ email, password }),
    }),

  login: (email: string, password: string) =>
    request<User>("/auth/login", {
      method: "POST",
      body: jsonBody({ email, password }),
    }),

  logout: () => request<void>("/auth/logout", { method: "POST" }),

  movies: () => request<Movie[]>("/movies"),

  movie: (id: string) => request<Movie>(`/movies/${encodeURIComponent(id)}`),

  showtimes: (movieId?: string) =>
    request<Showtime[]>(
      `/showtimes${movieId ? `?movie_id=${encodeURIComponent(movieId)}` : ""}`,
    ),

  showtime: (id: string) =>
    request<Showtime>(`/showtimes/${encodeURIComponent(id)}`),

  seats: (showtimeId: string) =>
    request<SeatsResponse>(`/showtimes/${encodeURIComponent(showtimeId)}/seats`),

  hold: (showtimeId: string, seatId: string) =>
    request<Hold>(
      `/showtimes/${encodeURIComponent(showtimeId)}/seats/${encodeURIComponent(seatId)}/hold`,
      { method: "POST" },
    ),

  newChallenge: () => request<Challenge>("/payment/challenge"),

  checkout: (sessionIds: string[], challengeId: string, answer: string) =>
    request<{ bookings: Booking[] }>("/checkout", {
      method: "POST",
      body: jsonBody({
        session_ids: sessionIds,
        payment: { id: challengeId, answer },
      }),
    }),

  release: (sessionId: string) =>
    request<void>(`/sessions/${encodeURIComponent(sessionId)}`, {
      method: "DELETE",
    }),

  myBookings: () => request<Booking[]>("/me/bookings"),

  cancelBooking: (bookingId: string) =>
    request<void>(`/bookings/${encodeURIComponent(bookingId)}`, {
      method: "DELETE",
    }),

  myTickets: () => request<TicketDetail[]>("/me/tickets"),

  notifications: () => request<AppNotification[]>("/me/notifications"),

  markNotificationsRead: () =>
    request<void>("/me/notifications/read", { method: "POST" }),

  createMovie: (movie: {
    id?: string;
    title: string;
    description: string;
    duration_min: number;
    poster_url: string;
  }) =>
    request<Movie>("/admin/movies", {
      method: "POST",
      body: jsonBody(movie),
    }),

  createShowtime: (showtime: {
    movie_id: string;
    starts_at: string;
    screen_name: string;
    rows: number;
    seats_per_row: number;
    premium_rows: number;
    premium_price_cents: number;
    standard_price_cents: number;
  }) =>
    request<Showtime>("/admin/showtimes", {
      method: "POST",
      body: jsonBody(showtime),
    }),
};

export function errorMessage(error: unknown): string {
  if (error instanceof ApiError) {
    return error.code.replaceAll("_", " ");
  }
  if (error instanceof Error) {
    return error.message;
  }
  return "something went wrong";
}
