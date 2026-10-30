import { create } from "zustand";
import { apiErrorFields, apiErrorMessage, isSdkError } from "../api/errors";
import {
  registerUser,
  resendEmail as resendEmailSdk,
  type UserAuth,
} from "../client";

interface SignUpPayload {
  name: string;
  email: string;
  username: string;
  password: string;
  email_marketing: boolean;
  sig?: string;
  device_code?: string;
}

/**
 * How a resend of the confirmation code ended: sent, refused because one went out within the last
 * minute, or failed for any other reason.
 */
export type ResendOutcome = "sent" | "throttled" | "failed";

interface SignUpState {
  signUpAuth: UserAuth | null;
  signUpLoading: boolean;
  signUpError: string | null;
  signUpServerFields: Record<string, string>;

  resendLoading: boolean;
  resendError: string | null;

  signUp: (payload: SignUpPayload) => Promise<string | null>;
  clearSignUpServerField: (field: string) => void;
  resetSignUpErrors: () => void;
  resendEmail: (identifier: string) => Promise<ResendOutcome>;
  resetResendError: () => void;
}

/**
 * The sign-up flow: the account request and the resend of its confirmation code. Field-level
 * errors from the server are kept apart from the general one, so each lands on its own input.
 */
export const useSignUpStore = create<SignUpState>()((set) => ({
  signUpAuth: null,
  signUpLoading: false,
  signUpError: null,
  signUpServerFields: {},

  resendLoading: false,
  resendError: null,

  signUp: async (payload) => {
    set({
      signUpLoading: true,
      signUpError: null,
      signUpServerFields: {},
      signUpAuth: null,
    });
    try {
      const { data } = await registerUser({
        body: payload,
        throwOnError: true,
      });
      const response = (data ?? {}) as Partial<UserAuth>;
      set({
        signUpLoading: false,
        signUpAuth: response.token ? (response as UserAuth) : null,
      });
      return response.token ?? null;
    } catch (error: unknown) {
      const fields = apiErrorFields(error);
      if (Object.keys(fields).length > 0) {
        set({ signUpLoading: false, signUpServerFields: fields });
        return null;
      }
      if (import.meta.env.DEV && isSdkError(error)) {
        console.warn("Unhandled sign-up error response:", {
          status: error.status,
        });
      }
      set({ signUpLoading: false, signUpError: apiErrorMessage(error) });
      return null;
    }
  },

  clearSignUpServerField: (field) =>
    set((s) => ({
      signUpServerFields: Object.fromEntries(
        Object.entries(s.signUpServerFields).filter(([f]) => f !== field),
      ),
    })),

  resetSignUpErrors: () => set({ signUpError: null, signUpServerFields: {} }),

  resendEmail: async (identifier) => {
    set({ resendLoading: true, resendError: null });
    try {
      await resendEmailSdk({ body: { identifier }, throwOnError: true });
      set({ resendLoading: false });
      return "sent";
    } catch (error: unknown) {
      const throttled = isSdkError(error) && error.status === 429;
      set({
        resendLoading: false,
        resendError: throttled
          ? "Wait a minute before asking for another email."
          : "Failed to resend email. Please try again.",
      });
      return throttled ? "throttled" : "failed";
    }
  },

  resetResendError: () => {
    set({ resendError: null });
  },
}));
