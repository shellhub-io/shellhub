import { useNavigate } from "react-router-dom";
import { useAuthStore } from "@/stores/authStore";

/**
 * Leaves for the sign-in page with a notice, handing it the sign-out of a session whose token the
 * server has just revoked. The token stays in the store until the sign-in page ends the session,
 * and a 401 that lands before then leaves the session alone instead of reloading the page and
 * losing the notice.
 */
export function useHandOverSignOut() {
  const navigate = useNavigate();

  return (notice: string) => {
    useAuthStore.setState({ signOutHandedOver: true });
    void navigate("/login", { state: { notice, signOut: true } });
  };
}
