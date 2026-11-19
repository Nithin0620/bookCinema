import { BrowserRouter, Navigate, Route, Routes } from "react-router-dom";
import { AuthProvider } from "./AuthContext";
import { Layout } from "./components/Layout";
import { AdminRoute, ProtectedRoute } from "./components/ProtectedRoute";
import { Admin } from "./pages/Admin";
import { AuthPage } from "./pages/AuthPage";
import { Movies } from "./pages/Movies";
import { MyBookings } from "./pages/MyBookings";
import { Notifications } from "./pages/Notifications";
import { Seats } from "./pages/Seats";
import { Tickets } from "./pages/Tickets";

export function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <Layout>
          <Routes>
            <Route path="/" element={<Movies />} />
            <Route path="/login" element={<AuthPage mode="login" />} />
            <Route path="/register" element={<AuthPage mode="register" />} />
            <Route
              path="/showtimes/:id"
              element={
                <ProtectedRoute>
                  <Seats />
                </ProtectedRoute>
              }
            />
            <Route
              path="/bookings"
              element={
                <ProtectedRoute>
                  <MyBookings />
                </ProtectedRoute>
              }
            />
            <Route
              path="/tickets"
              element={
                <ProtectedRoute>
                  <Tickets />
                </ProtectedRoute>
              }
            />
            <Route
              path="/notifications"
              element={
                <ProtectedRoute>
                  <Notifications />
                </ProtectedRoute>
              }
            />
            <Route
              path="/admin"
              element={
                <AdminRoute>
                  <Admin />
                </AdminRoute>
              }
            />
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </Layout>
      </AuthProvider>
    </BrowserRouter>
  );
}
