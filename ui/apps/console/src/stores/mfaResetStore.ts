import { create } from "zustand";
import { isSdkError } from "../api/errors";
import { requestResetMfa, resetMfa } from "../client";
import { useAuthStore } from "./authStore";

const mfaResetThrottledMessage =
  "A reset was already requested in the last minute. Check both inboxes, or wait a minute and try again.";

/** mfaResetRequestFailedMessage is shown when a reset request fails for any reason but a 429. */
export const mfaResetRequestFailedMessage = "Unable to send reset emails. Please try again.";

/**
 * mfaResetCodesRejectedMessage is shown when completing a reset fails for any reason, a refused
 * code pair or otherwise. The server cancels the codes after three pairs with one right code, so
 * the text points the user to a new reset.
 */
export const mfaResetCodesRejectedMessage =
  "Invalid verification codes. Three wrong attempts cancel the codes, so request a new reset if this keeps failing.";

function mfaResetRequestError(err: unknown): string {
  if (isSdkError(err) && err.status === 429) return mfaResetThrottledMessage;

  return mfaResetRequestFailedMessage;
}

/** What the sign-in page says after an MFA reset, which signs no one in. */
export const mfaResetNotice = "MFA is off. Sign in with your password.";

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
      const { data } = await requestResetMfa({
        body: { identifier },
        throwOnError: true,
      });
      set({
        mfaResetToken: data.token,
        mfaResetIdentifier: identifier,
        loading: false,
      });
    } catch (err) {
      set({
        loading: false,
        error: mfaResetRequestError(err),
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
      await resetMfa({
        path: { "user-id": mfaResetToken },
        body: {
          main_email_code: mainEmailCode,
          recovery_email_code: recoveryEmailCode,
        },
        throwOnError: true,
      });

      useAuthStore.setState({ mfaToken: null });
      set({ mfaResetToken: null, mfaResetIdentifier: null, loading: false });
    } catch {
      set({
        loading: false,
        error: mfaResetCodesRejectedMessage,
      });
      throw new Error("Invalid codes");
    }
  },

  reset: () => set(initialState),
}));
