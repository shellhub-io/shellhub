import { create } from "zustand";
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
  name: string | null;
  isLoggedIn: boolean;
  loading: boolean;
  error: string | null;
  login: (username: string, password: string) => Promise<void>;
  logout: () => void;
  fetchUser: () => Promise<void>;
  restoreSession: () => void;
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

export const useAuthStore = create<AuthState>((set, get) => ({
  token: null,
  user: null,
  userId: null,
  email: null,
  username: null,
  recoveryEmail: null,
  tenant: null,
  name: null,
  isLoggedIn: false,
  loading: false,
  error: null,

  login: async (username: string, password: string) => {
    set({ loading: true, error: null });
    try {
      const data = await apiLogin({ username, password });
      localStorage.setItem("token", data.token);
      localStorage.setItem("user", data.user);
      localStorage.setItem("tenant", data.tenant);
      localStorage.setItem("email", data.email);
      localStorage.setItem("userId", data.id);
      set({
        token: data.token,
        user: data.user,
        userId: data.id,
        email: data.email,
        tenant: data.tenant,
        name: data.name,
        isLoggedIn: true,
        loading: false,
      });
    } catch {
      set({ loading: false, error: "Invalid username or password" });
    }
  },

  logout: () => {
    localStorage.removeItem("token");
    localStorage.removeItem("user");
    localStorage.removeItem("tenant");
    localStorage.removeItem("email");
    localStorage.removeItem("userId");
    set({
      token: null,
      user: null,
      userId: null,
      email: null,
      username: null,
      recoveryEmail: null,
      tenant: null,
      name: null,
      isLoggedIn: false,
    });
  },

  fetchUser: async () => {
    try {
      const user = await getAuthUser();
      localStorage.setItem("userId", user.id);
      localStorage.setItem("email", user.email);
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

  restoreSession: () => {
    const token = localStorage.getItem("token");
    const user = localStorage.getItem("user");
    const userId = localStorage.getItem("userId");
    const tenant = localStorage.getItem("tenant");
    const email = localStorage.getItem("email");
    if (token) {
      set({ token, user, userId, tenant, email, isLoggedIn: true });
    }
  },

  updateProfile: async (data) => {
    await updateUser(data);
    await get().fetchUser();
  },

  updatePassword: async (currentPassword, newPassword) => {
    await apiUpdatePassword(currentPassword, newPassword);
  },
}));
