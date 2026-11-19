export interface User {
  id: string;
  email: string;
  created_at: string;
  is_admin: boolean;
}

export interface Movie {
  id: string;
  title: string;
  description: string;
  duration_min: number;
  poster_url: string;
}

export interface Showtime {
  id: string;
  movie_id: string;
  starts_at: string;
  screen_name: string;
  rows: number;
  seats_per_row: number;
  premium_rows: number;
  premium_price_cents: number;
  standard_price_cents: number;
}

export interface SeatView {
  seat_id: string;
  section: "premium" | "standard";
  price_cents: number;
  status: "available" | "held" | "confirmed";
  mine: boolean;
  session_id?: string;
  expires_at?: string;
}

export interface SeatsResponse {
  showtime: Showtime;
  seats: SeatView[];
}

export interface Hold {
  id: string;
  showtime_id: string;
  seat_id: string;
  user_id: string;
  expires_at: string;
}

export interface Booking {
  id: string;
  showtime_id: string;
  seat_id: string;
  user_id: string;
  price_cents: number;
  created_at: string;
}

export interface Challenge {
  id: string;
  question: string;
  expires_at: string;
}

export interface TicketDetail {
  ticket: {
    id: string;
    booking_id: string;
    code: string;
    created_at: string;
  };
  booking: Booking;
}

export interface AppNotification {
  id: string;
  user_id: string;
  title: string;
  body: string;
  read_at: string | null;
  created_at: string;
}
