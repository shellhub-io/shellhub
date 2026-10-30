import { create } from "zustand";
import { persist, createJSONStorage } from "zustand/middleware";
import { type Role } from "../utils/permission";
import {
  login as loginSdk,
  getUserInfo,
  updateUser as updateUserSdk,
  deleteUser as deleteUserSdk,
  authMfa,
  mfaRecover,
  confirmAccount as confirmAccountSdk,
  type UserOnboardingSurvey,
  type UserOrigin,
} from "../client";
import { queryClient } from "../api/queryClient";
import { isSdkError } from "../api/errors";
import { tearDownChatwoot } from "../hooks/chatwootRuntime";
import { useVaultStore } from "./vaultStore";
import { useTerminalStore } from "./terminalStore";
import { useWorkspaceTabsStore } from "./workspaceTabsStore";
import { useRecentDevicesStore } from "./recentDevicesStore";
import { setRecordingsScope } from "../utils/recordings";
import {
  authSessionKey,
  authSessionVersion,
  signedInSession,
} from "./authSession";

interface AuthState {
  token: string | null;
  user: string | null;
  userId: string | null;
  email: string | null;
  username: string | null;
  origin: UserOrigin | null;
  recoveryEmail: string | null;
  tenant: string | null;
  role: Role | null;
  isAdmin: boolean;
  name: string | null;
  loading: boolean;
  error: string | null;
  mfaEnabled: boolean;
  mfaToken: string | null;
  mfaRecoveryExpiry: number | null;
  onboardingSurvey: UserOnboardingSurvey | null;
  impersonated: boolean;
  signOutHandedOver: boolean;
  login: (username: string, password: string) => Promise<void>;
  loginWithToken: (token: string) => Promise<void>;
  logout: () => void;
  fetchUser: () => Promise<void>;
  setSession: (data: {
    token: string;
    tenant: string;
    role?: Role | null;
  }) => void;
  updateProfile: (data: {
    name?: string;
    username?: string;
    email?: string;
    recovery_email?: string;
  }) => Promise<void>;
  updatePassword: (
    currentPassword: string,
    newPassword: string,
  ) => Promise<void>;
  deleteUser: () => Promise<void>;
  loginWithMfa: (code: string) => Promise<void>;
  confirmAccount: (identifier: string, code: string) => Promise<boolean>;
  recoverWithCode: (code: string, identifier?: string) => Promise<void>;
  updateMfaStatus: (enabled: boolean) => void;
  setMfaToken: (token: string) => void;
  setOnboardingSurvey: (outcome: UserOnboardingSurvey) => void;
  isFromAbandonedSession: (sentToken: string | null) => boolean;
}

const initialState = {
  token: null,
  user: null,
  userId: null,
  email: null,
  username: null,
  origin: null,
  recoveryEmail: null,
  tenant: null,
  role: null,
  isAdmin: false,
  name: null,
  loading: false,
  error: null,
  mfaEnabled: false,
  mfaToken: null,
  mfaRecoveryExpiry: null,
  onboardingSurvey: null,
  impersonated: false,
  signOutHandedOver: false,
};

/**
 * Who is signed in, and the transitions in and out of that state — sign-in, MFA, namespace
 * switching, sign-out. Persisted, so a reload does not sign the user out; the token is the
 * part that matters and everything else is derived from it.
 *
 * Every request is from an abandoned session from the moment its sign-out is handed over to the
 * sign-in page until logout, or until the hand-over is withdrawn. Otherwise, only a request that
 * carried a token the store no longer holds is; one that carried none is not.
 */
export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      ...initialState,

      login: async (username: string, password: string) => {
        set({ loading: true, mfaToken: null });
        try {
          const { data, response } = await loginSdk({
            body: { username, password },
            throwOnError: true,
          });

          const mfaToken = response.headers.get("x-mfa-token");

          if (mfaToken) {
            set({
              mfaToken,
              mfaEnabled: true,
              user: username,
              loading: false,
            });
            return;
          }

          set({ ...signedInSession(data), loading: false });
        } catch (err) {
          const currentState = get();
          if (currentState.mfaToken) {
            set({ loading: false, user: username, mfaEnabled: true });
            return;
          }
          set({ loading: false });
          throw err;
        }
      },

      loginWithToken: async (token: string) => {
        set({ loading: true, token });
        try {
          const { data } = await getUserInfo({ throwOnError: true });
          set({
            ...signedInSession(data),
            token,
            origin: data.origin ?? null,
            loading: false,
          });
        } catch {
          tearDownChatwoot("logout");
          set({ ...initialState });
          throw new Error("Token login failed");
        }
      },

      logout: () => {
        tearDownChatwoot("logout");
        useVaultStore.getState().lock();
        const terminal = useTerminalStore.getState();
        terminal.sessions.forEach((s) => terminal.close(s.id));
        terminal.recordings.forEach((r) => terminal.closeRecording(r.id));
        useWorkspaceTabsStore.getState().clear();
        useRecentDevicesStore.getState().clear();
        set(initialState);
        localStorage.removeItem("shellhub-session");
        queryClient.clear();
      },

      fetchUser: async () => {
        try {
          const { data } = await getUserInfo({ throwOnError: true });
          const user = data;
          set({
            user: user.user,
            username: user.user,
            userId: user.id,
            email: user.email,
            origin: user.origin ?? null,
            recoveryEmail: user.recovery_email,
            name: user.name,
            tenant: user.tenant,
            isAdmin: user.admin ?? false,
            mfaEnabled: user.mfa || false,
          });
        } catch {
          set({ isAdmin: false });
        }
      },

      setSession: ({ token, tenant, role }) => {
        set({ token, tenant, role: role === undefined ? get().role : role });
      },

      updateProfile: async (data) => {
        await updateUserSdk({ body: data, throwOnError: true });
        await get().fetchUser();
      },

      updatePassword: async (currentPassword, newPassword) => {
        await updateUserSdk({
          body: { current_password: currentPassword, password: newPassword },
          throwOnError: true,
        });
      },

      deleteUser: async () => {
        await deleteUserSdk({ throwOnError: true });
        get().logout();
        window.location.replace("/login");
      },

      confirmAccount: async (identifier, code) => {
        const { data, response } = await confirmAccountSdk({
          body: { identifier, code },
          throwOnError: true,
        });
        if (response.status === 204 || !data) return false;
        set(signedInSession(data));
        return true;
      },

      loginWithMfa: async (code: string) => {
        const { mfaToken } = get();
        if (!mfaToken) {
          throw new Error("No MFA token available");
        }

        set({ loading: true, error: null });
        try {
          const { data } = await authMfa({
            body: { token: mfaToken, code },
            throwOnError: true,
          });
          set({
            ...signedInSession(data),
            mfaToken: null, // Clear temporary token
            mfaEnabled: true,
            loading: false,
          });
        } catch (err) {
          if (isSdkError(err) && err.status === 429) {
            set({
              loading: false,
              error:
                "Too many failed attempts. Please wait before trying again.",
            });
            throw err;
          }
          set({ loading: false, error: "Invalid verification code" });
          throw new Error("Invalid verification code");
        }
      },

      recoverWithCode: async (code: string, identifier?: string) => {
        const username = identifier || get().user || get().username;
        if (!username) {
          set({ error: "Username or email is required" });
          throw new Error("Username or email is required");
        }

        set({ loading: true, error: null });
        try {
          const { data, response } = await mfaRecover({
            body: { identifier: username, recovery_code: code },
            throwOnError: true,
          });

          const expiresAt = response.headers.get("x-expires-at") || "";

          let expiryValue: number | null = null;
          if (expiresAt) {
            const parsed = parseInt(expiresAt, 10);
            expiryValue = !isNaN(parsed) ? parsed : null;
          }

          set({
            ...signedInSession(data),
            mfaEnabled: true,
            mfaToken: null,
            mfaRecoveryExpiry: expiryValue,
            loading: false,
          });
        } catch {
          set({ loading: false, error: "Invalid recovery code or username" });
          throw new Error("Invalid recovery code or username");
        }
      },

      updateMfaStatus: (enabled: boolean) => {
        set({ mfaEnabled: enabled });
      },

      setMfaToken: (token: string) => {
        set({ mfaToken: token });
      },

      setOnboardingSurvey: (outcome) => {
        set({ onboardingSurvey: outcome });
      },

      isFromAbandonedSession: (sentToken: string | null) => {
        const { token, signOutHandedOver } = get();
        if (signOutHandedOver) return true;
        return sentToken !== null && sentToken !== token;
      },
    }),
    {
      name: authSessionKey,
      version: authSessionVersion,
      storage: createJSONStorage(() => localStorage),
      partialize: (state) => ({
        token: state.token,
        user: state.user,
        userId: state.userId,
        email: state.email,
        origin: state.origin,
        tenant: state.tenant,
        role: state.role,
        isAdmin: state.isAdmin,
        name: state.name,
        mfaEnabled: state.mfaEnabled,
        onboardingSurvey: state.onboardingSurvey,
        impersonated: state.impersonated,
      }),
    },
  ),
);

setRecordingsScope(useAuthStore.getState().userId);
useAuthStore.subscribe((state, prev) => {
  if (state.userId !== prev.userId) setRecordingsScope(state.userId);
});

/**
 * How the signed-in user is named across the console: the username, else the display name, else
 * the email, else "Account". A selector for useAuthStore.
 */
export const accountDisplayName = (
  s: Pick<AuthState, "user" | "name" | "email">,
) => s.user || s.name || s.email || "Account";
