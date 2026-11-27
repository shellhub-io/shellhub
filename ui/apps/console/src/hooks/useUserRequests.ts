import { useMutation, useQuery } from "@tanstack/react-query";
import { useNavigate } from "react-router-dom";
import { getUserInfoOptions, resetMfaMutation } from "@/client";
import { features } from "@/edition";
import { useAuthStore } from "@/stores/authStore";

/**
 * The signed-in account's info, namespace allowance included. Community never requests it, so
 * `data` stays undefined there.
 */
export function useUserInfo() {
  return useQuery({
    ...getUserInfoOptions(),
    enabled: features().paid,
  });
}

/**
 * useCompleteMfaReset finishes an MFA reset with the codes from both emails, for the reset the path
 * id names. On success it drops authStore's mfaToken, since MFA is now off and the half-finished
 * sign-in that token stands for can no longer complete, and sends the user to sign in with the
 * password, signing out whoever is signed in on this browser. The sign-in page replaces the reset
 * page in history, so going back cannot reopen a reset that is used. A failure leaves the user on
 * the page, with the API error, or the fetch failure, as the mutation's error, and rejects only
 * through mutateAsync.
 */
export function useCompleteMfaReset() {
  const navigate = useNavigate();

  return useMutation({
    ...resetMfaMutation(),
    onSuccess: () => {
      useAuthStore.setState({ mfaToken: null });
      void navigate("/login", {
        replace: true,
        state: {
          notice: "MFA is off. Sign in with your password.",
          signOut: true,
        },
      });
    },
  });
}
