import type { ReactNode } from "react";
import { Navigate, useLocation } from "react-router-dom";
import { useAuth } from "../AuthContext";

export function ProtectedRoute({ children }: { children: ReactNode }) {
  const { user, loading } = useAuth();
  const location = useLocation();

  if (loading) {
    return <div className="state">Loading...</div>;
  }
  if (!user) {
    return <Navigate to="/login" state={{ from: location.pathname }} replace />;
  }
  return <>{children}</>;
}

export function AdminRoute({ children }: { children: ReactNode }) {
  const { user, loading } = useAuth();

  if (loading) {
    return <div className="state">Loading...</div>;
  }
  if (!user) {
    return <Navigate to="/login" state={{ from: "/admin" }} replace />;
  }
  if (!user.is_admin) {
    return <Navigate to="/" replace />;
  }
  return <>{children}</>;
}
