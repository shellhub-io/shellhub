import { useNavigate } from "react-router-dom";
import { useAuthStore } from "@/stores/authStore";

/**
 * Runs revoke, a request that makes the server revoke the session's token, then leaves for the
 * sign-in page with a notice, handing it the sign-out. The session counts as abandoned from
 * before revoke is sent, so a 401 that any other request meets once the server has revoked the
 * token leaves the session alone instead of reloading the page and losing the notice. The token
 * stays in the store until the sign-in page ends the session. When revoke rejects, the hand-over
 * is withdrawn, the page stays where it is and the returned promise rejects with revoke's error.
 * That includes a 401 on revoke itself, from a token the server revoked elsewhere: it reaches the
 * caller as an error instead of ending the session, and the next request that meets it ends it.
 */
export function useHandOverSignOut() {
  const navigate = useNavigate();

  return async (revoke: () => Promise<unknown>, notice: string) => {
    useAuthStore.setState({ signOutHandedOver: true });
    try {
      await revoke();
      void navigate("/login", { state: { notice, signOut: true } });
    } catch (err) {
      useAuthStore.setState({ signOutHandedOver: false });
      throw err;
    }
  };
}
