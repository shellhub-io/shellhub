import { create } from "zustand";
import { persist, createJSONStorage } from "zustand/middleware";
import axios from "axios";
import {
  login as apiLogin,
  getAuthUser,
  updateUser,
  updatePassword as apiUpdatePassword,
} from "../api/auth";

interface AuthState {
  token: string | null;
  user: string | null;
  userId: string | null;
  email: string | null;
  username: string | null;
  recoveryEmail: string | null;
  tenant: string | null;
  role: string | null;
  name: string | null;
  loading: boolean;
  error: string | null;
  login: (username: string, password: string) => Promise<void>;
  logout: () => void;
  fetchUser: () => Promise<void>;
  setSession: (data: { token: string; tenant: string; role?: string }) => void;
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
}

const initialState = {
  token: null,
  user: null,
  userId: null,
  email: null,
  username: null,
  recoveryEmail: null,
  tenant: null,
  role: null,
  name: null,
  loading: false,
  error: null,
};

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      ...initialState,

      login: async (username: string, password: string) => {
        set({ loading: true, error: null });
        try {
          const data = await apiLogin({ username, password });
          set({
            token: data.token,
            user: data.user,
            userId: data.id,
            email: data.email,
            tenant: data.tenant,
            name: data.name,
            loading: false,
          });
        } catch (err) {
          const status = axios.isAxiosError(err) ? err.response?.status : null;
          const error = status === 403
            ? "Your account has not been confirmed. Please check your email for the activation link."
            : "Invalid username or password";
          set({ loading: false, error, token: null });
        }
      },

      logout: () => {
        set(initialState);
      },

      fetchUser: async () => {
        try {
          const user = await getAuthUser();
          set({
            user: user.user,
            username: user.user,
            userId: user.id,
            email: user.email,
            recoveryEmail: user.recovery_email,
            name: user.name,
            tenant: user.tenant,
          });
        } catch {
          /* session expired — interceptor handles redirect */
        }
      },

      setSession: ({ token, tenant, role }) => {
        set({ token, tenant, role: role ?? get().role });
      },

      updateProfile: async (data) => {
        await updateUser(data);
        await get().fetchUser();
      },

      updatePassword: async (currentPassword, newPassword) => {
        await apiUpdatePassword(currentPassword, newPassword);
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
        tenant: state.tenant,
        role: state.role,
        name: state.name,
      }),
    },
  ),
);
