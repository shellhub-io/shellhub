import { create } from "zustand";
import { requestResetMfa, resetMfa } from "../client";
import { useAuthStore } from "./authStore";

interface MfaResetState {
  mfaResetUserId: string | null;
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
  mfaResetUserId: null,
  mfaResetIdentifier: null,
  loading: false,
  error: null,
};

export const useMfaResetStore = create<MfaResetState>()((set, get) => ({
  ...initialState,

  requestMfaReset: async (identifier: string) => {
    set({ loading: true, error: null });
    try {
      await requestResetMfa({
        body: { identifier },
        throwOnError: true,
      });
      set({
        mfaResetUserId: identifier,
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
    const { mfaResetUserId } = get();
    if (!mfaResetUserId) {
      set({ error: "Invalid reset session. Please start over." });
      throw new Error("No user ID available");
    }

    set({ loading: true, error: null });
    try {
      const { data } = await resetMfa({
        path: { "user-id": mfaResetUserId },
        body: {
          main_email_code: mainEmailCode,
          recovery_email_code: recoveryEmailCode,
        },
        throwOnError: true,
      });

      const userData = data;
      useAuthStore.setState({
        token: userData.token,
        user: userData.user,
        userId: userData.id,
        email: userData.email,
        tenant: userData.tenant,
        name: userData.name,
        isAdmin: userData.admin ?? false,
        mfaEnabled: userData.mfa || false,
      });
      set({ mfaResetUserId: null, mfaResetIdentifier: null, loading: false });
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
