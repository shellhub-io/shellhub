import { Navigate, useParams } from "react-router-dom";

/**
 * Sends a link to the account pages' old /profile address to the same page under /account. It
 * sits outside the namespace guard, so a user without a namespace following an old link still
 * reaches their account rather than the create-namespace screen.
 */
export default function LegacyProfileRedirect() {
  const rest = useParams()["*"];
  return <Navigate to={rest ? `/account/${rest}` : "/account"} replace />;
}
