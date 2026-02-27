import { create } from "zustand";
import axios from "axios";
import {
  signUp as apiSignUp,
  resendEmail as apiResendEmail,
  validateAccount as apiValidateAccount,
  type SignUpPayload,
} from "../api/users";

export type ValidationStatus = "idle" | "processing" | "success" | "failed" | "failed-token";

interface SignUpState {
  signUpToken: string | null;
  signUpTenant: string | null;
  signUpLoading: boolean;
  signUpError: string | null;
  signUpServerFields: string[]; // raw server field names from 400/409 responses

  resendLoading: boolean;
  resendError: string | null;

  validationStatus: ValidationStatus;

  signUp: (payload: SignUpPayload) => Promise<string | null>;
  clearSignUpServerField: (field: string) => void;
  resetSignUpErrors: () => void;
  resendEmail: (username: string) => Promise<boolean>;
  validateAccount: (email: string, token: string, signal?: AbortSignal) => Promise<void>;
  resetValidation: () => void;
  setValidationFailed: () => void;
  resetResendError: () => void;
}

export const useSignUpStore = create<SignUpState>()((set) => ({
  signUpToken: null,
  signUpTenant: null,
  signUpLoading: false,
  signUpError: null,
  signUpServerFields: [],

  resendLoading: false,
  resendError: null,

  validationStatus: "idle",

  signUp: async (payload) => {
    set({ signUpLoading: true, signUpError: null, signUpServerFields: [], signUpToken: null, signUpTenant: null });
    try {
      const response = await apiSignUp(payload);
      set({
        signUpLoading: false,
        signUpToken: response.token ?? null,
        signUpTenant: response.tenant ?? null,
      });
      return response.token ?? null;
    } catch (error: unknown) {
      if (axios.isAxiosError(error)) {
        const status = error.response?.status;
        const data = error.response?.data;
        if ((status === 400 || status === 409) && Array.isArray(data)) {
          const fields = data.filter((f): f is string => typeof f === "string");
          set({ signUpLoading: false, signUpServerFields: fields });
          return null;
        }
        if (import.meta.env.DEV) {
          console.warn("Unhandled sign-up error response:", { status, data });
        }
      }
      set({ signUpLoading: false, signUpError: "An error occurred. Please try again." });
      return null;
    }
  },

  clearSignUpServerField: (field) =>
    set((s) => ({ signUpServerFields: s.signUpServerFields.filter((f) => f !== field) })),

  resetSignUpErrors: () => set({ signUpError: null, signUpServerFields: [] }),

  resendEmail: async (username) => {
    set({ resendLoading: true, resendError: null });
    try {
      await apiResendEmail(username);
      set({ resendLoading: false });
      return true;
    } catch {
      set({ resendLoading: false, resendError: "Failed to resend email. Please try again." });
      return false;
    }
  },

  validateAccount: async (email, token, signal) => {
    set({ validationStatus: "processing" });
    try {
      await apiValidateAccount(email, token, signal);
      set({ validationStatus: "success" });
    } catch (error: unknown) {
      // Ignore aborted requests (AbortController cleanup in Strict Mode / unmount).
      if (axios.isCancel(error)) return;
      const status = axios.isAxiosError(error) ? error.response?.status : null;
      // 400 = expired/invalid token, 401 = wrong token; both show "token" guidance.
      // 404 = user not found (wrong email); falls through to generic "failed".
      set({ validationStatus: (status === 400 || status === 401) ? "failed-token" : "failed" });
    }
  },

  resetValidation: () => set({ validationStatus: "idle" }),
  setValidationFailed: () => set({ validationStatus: "failed" }),

  resetResendError: () => {
    set({ resendError: null });
  },
}));
