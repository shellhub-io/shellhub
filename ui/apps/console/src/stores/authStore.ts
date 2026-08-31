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
  type UserOrigin,
} from "../client";
import { queryClient } from "../api/queryClient";
import { tearDownChatwoot } from "../hooks/chatwootRuntime";
import { useVaultStore } from "./vaultStore";
import { useTerminalStore } from "./terminalStore";
import { setRecordingsScope } from "../utils/recordings";

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
  login: (username: string, password: string) => Promise<void>;
  loginWithToken: (token: string) => Promise<void>;
  logout: () => void;
  fetchUser: () => Promise<void>;
  setSession: (data: { token: string; tenant: string; role?: Role }) => void;
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
  recoverWithCode: (code: string, identifier?: string) => Promise<void>;
  updateMfaStatus: (enabled: boolean) => void;
  setMfaToken: (token: string) => void;
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
};

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

          const userData = data;
          set({
            token: userData.token,
            user: userData.user,
            userId: userData.id,
            email: userData.email,
            tenant: userData.tenant,
            name: userData.name,
            isAdmin: userData.admin ?? false,
            mfaEnabled: userData.mfa || false,
            loading: false,
          });
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
            user: data.user,
            userId: data.id,
            email: data.email,
            origin: data.origin ?? null,
            tenant: data.tenant,
            name: data.name,
            isAdmin: data.admin ?? false,
            mfaEnabled: data.mfa || false,
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
        set({ token, tenant, role: role ?? get().role });
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
            token: data.token,
            user: data.user,
            userId: data.id,
            email: data.email,
            tenant: data.tenant,
            name: data.name,
            isAdmin: data.admin ?? false,
            mfaToken: null, // Clear temporary token
            mfaEnabled: true,
            loading: false,
          });
        } catch {
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

          const userData = data;
          const expiresAt = response.headers.get("x-expires-at") || "";

          let expiryValue: number | null = null;
          if (expiresAt) {
            const parsed = parseInt(expiresAt, 10);
            expiryValue = !isNaN(parsed) ? parsed : null;
          }

          set({
            token: userData.token,
            user: userData.user,
            userId: userData.id,
            email: userData.email,
            tenant: userData.tenant,
            name: userData.name,
            isAdmin: userData.admin ?? false,
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
    }),
    {
      name: "shellhub-session",
      version: 1,
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
      }),
    },
  ),
);

setRecordingsScope(useAuthStore.getState().userId);
useAuthStore.subscribe((state, prev) => {
  if (state.userId !== prev.userId) setRecordingsScope(state.userId);
});
