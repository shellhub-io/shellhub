import { create } from "zustand";
import { requestResetMFA, resetMFA } from "@/client/api";
import { useAuthStore } from "./authStore";

interface MfaResetState {
  mfaResetToken: string | null;
  mfaResetIdentifier: string | null;
  loading: boolean;
  error: string | null;
  requestMfaReset: (identifier: string) => Promise<void>;
  completeMfaReset: (
    mainEmailCode: string,
    recoveryEmailCode: string,
  ) => Promise<void>;
  reset: () => void;
}

const initialState = {
  mfaResetToken: null,
  mfaResetIdentifier: null,
  loading: false,
  error: null,
};

/**
 * The MFA reset flow, which spans three screens and an emailed code. It is a store rather than
 * route state because the user leaves the app for their mail and comes back.
 */
export const useMfaResetStore = create<MfaResetState>()((set, get) => ({
  ...initialState,

  requestMfaReset: async (identifier: string) => {
    set({ loading: true, error: null });
    try {
      const data = await requestResetMFA({ identifier });
      set({
        mfaResetToken: data.token,
        mfaResetIdentifier: identifier,
        loading: false,
      });
    } catch {
      set({
        loading: false,
        error: "Unable to send reset emails. Please check your identifier.",
      });
      throw new Error("Reset request failed");
    }
  },

  completeMfaReset: async (
    mainEmailCode: string,
    recoveryEmailCode: string,
  ) => {
    const { mfaResetToken } = get();
    if (!mfaResetToken) {
      set({ error: "Invalid reset session. Please start over." });
      throw new Error("No reset token available");
    }

    set({ loading: true, error: null });
    try {
      const data = await resetMFA(mfaResetToken, {
        main_email_code: mainEmailCode,
        recovery_email_code: recoveryEmailCode,
      });

      useAuthStore.setState({
        token: data.token,
        user: data.user,
        userId: data.id,
        email: data.email,
        tenant: data.tenant,
        name: data.name,
        isAdmin: data.admin ?? false,
        mfaEnabled: data.mfa ?? false,
      });
      set({ mfaResetToken: null, mfaResetIdentifier: null, loading: false });
    } catch {
      set({
        loading: false,
        error: "Invalid verification codes. Please check and try again.",
      });
      throw new Error("Invalid codes");
    }
  },

  reset: () => set(initialState),
}));
