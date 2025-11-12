import { AxiosError } from "axios";
import { defineStore } from "pinia";
import { ref, computed } from "vue";
import * as authApi from "../api/auth";
import * as namespaceApi from "../api/namespaces";
import {
  configuration as apiConfiguration,
  reloadConfiguration as reloadApiConfiguration,
} from "@/api/http";
import { IMfaGenerate, IUserLogin, IMfaDisable, IMfaEnable, IMfaReset } from "@/interfaces/IUserLogin";
import { IUser } from "@/interfaces/IUser";

interface IAuthData {
  token?: string;
  user?: string;
  username?: string;
  name?: string;
  tenantId?: string;
  tenant?: string;
  email?: string;
  id?: string;
  role?: string;
  recovery_email?: string;
  mfa?: boolean;
  auth_methods?: string[];
  admin?: boolean;
}

const useAuthStore = defineStore("auth", () => {
  const token = ref(localStorage.getItem("token") || "");
  const username = ref(localStorage.getItem("user") || "");
  const name = ref(localStorage.getItem("name") || "");
  const tenantId = ref(localStorage.getItem("tenant") || "");
  const email = ref(localStorage.getItem("email") || "");
  const id = ref(localStorage.getItem("id") || "");
  const role = ref(localStorage.getItem("role") || "");
  const isAdmin = ref(localStorage.getItem("admin") === "true");
  const recoveryEmail = ref("");
  const isMfaEnabled = ref(false);
  const recoveryCode = ref("");
  const authMethods = ref(["local"]);
  const isRecoveringMfa = ref(false);
  const loginTimeout = ref(0);
  const disableTimeout = ref(0);
  const mfaToken = ref<string | undefined>();

  const isLoggedIn = computed(() => !!token.value);
  const showForceRecoveryMail = computed(() => !recoveryEmail.value && isMfaEnabled.value);
  const showRecoveryModal = computed(() => isRecoveringMfa.value && isMfaEnabled.value);

  const persistAuth = (data: IAuthData) => {
    token.value = data.token || "";
    username.value = data.user || data.username || "";
    name.value = data.name || "";
    tenantId.value = data.tenantId || data.tenant || "";
    email.value = data.email || "";
    id.value = data.id || "";
    role.value = data.role || "";
    isAdmin.value = data.admin || false;
    recoveryEmail.value = data.recovery_email || "";
    isMfaEnabled.value = data.mfa || false;
    authMethods.value = data.auth_methods || ["local"];

    localStorage.setItem("token", token.value);
    localStorage.setItem("user", username.value);
    localStorage.setItem("name", name.value);
    localStorage.setItem("tenant", tenantId.value);
    localStorage.setItem("email", email.value);
    localStorage.setItem("id", id.value);
    localStorage.setItem("role", role.value);
    localStorage.setItem("admin", String(isAdmin.value));
    localStorage.setItem("recovery_email", recoveryEmail.value);
    localStorage.setItem("mfa", String(isMfaEnabled.value));
    localStorage.setItem("namespacesWelcome", JSON.stringify({}));

    apiConfiguration.accessToken = token.value;
    reloadApiConfiguration();
  };

  const login = async (userInput: IUserLogin) => {
    try {
      const resp = await authApi.login(userInput);
      persistAuth(resp.data);
    } catch (error) {
      const axiosError = error as AxiosError;
      const token = axiosError.response?.headers["x-mfa-token"] as string;
      if (token) {
        isMfaEnabled.value = true;
        localStorage.setItem("mfa", "true");
        mfaToken.value = token;
        return;
      }

      loginTimeout.value = axiosError?.response?.headers["x-account-lockout"] as number;
      throw error;
    }
  };

  const loginWithToken = async (tokenInput: string) => {
    apiConfiguration.accessToken = tokenInput;
    reloadApiConfiguration();
    const resp = await authApi.getUserInfo();
    persistAuth(resp.data);
  };

  const validateMfa = async (verificationCode: string) => {
    const resp = await authApi.validateMFA({ token: mfaToken.value as string, code: verificationCode });
    if (resp.status === 200) {
      persistAuth(resp.data);
      mfaToken.value = resp.data.token;
      isMfaEnabled.value = true;
    }
  };

  const recoverMfa = async (code: string) => {
    name.value = localStorage.getItem("name") || "";
    const resp = await authApi.recoverMfa({ identifier: name.value, recovery_code: code });
    if (resp.status === 200) {
      persistAuth(resp.data);
      mfaToken.value = resp.data.token;
      isRecoveringMfa.value = true;
      recoveryCode.value = code;
      disableTimeout.value = resp.headers["x-expires-at"] as number;
    }
  };

  const disableMfa = async (data: Partial<IMfaDisable>) => {
    await authApi.disableMfa(data);
    isMfaEnabled.value = false;
    localStorage.setItem("mfa", "false");
  };

  const generateMfa = async () => {
    const { data } = await authApi.generateMfa();
    return data as IMfaGenerate;
  };

  const getUserInfo = async () => {
    const resp = await authApi.getUserInfo();
    if (resp.status === 200) persistAuth(resp.data);
  };

  const enterInvitedNamespace = async (tenantId: string) => {
    localStorage.removeItem("role");
    const res = await namespaceApi.switchNamespace(tenantId);
    if (res.status === 200) {
      persistAuth({ ...res.data, tenantId });
    }
  };

  const requestMfaReset = async () => {
    name.value = localStorage.getItem("name") || "";
    await authApi.requestResetMfa(name.value);
  };

  const resetMfa = async (data: IMfaReset) => {
    const resp = await authApi.resetMfa(data);
    if (resp.status === 200) persistAuth(resp.data);
  };

  const enableMfa = async (data: IMfaEnable) => {
    const resp = await authApi.enableMFA(data);
    if (resp.status === 200) isMfaEnabled.value = true;
  };

  const logout = () => {
    token.value = "";
    username.value = "";
    name.value = "";
    tenantId.value = "";
    email.value = "";
    id.value = "";
    role.value = "";
    isMfaEnabled.value = false;
    mfaToken.value = "";

    [
      "token", "user", "tenant", "namespacesWelcome", "noNamespace",
      "email", "id", "name", "role", "mfa", "recovery_email",
    ].forEach((key) => localStorage.removeItem(key));
  };

  const deleteUser = async () => {
    await authApi.deleteUser();
    logout();
  };

  const updateUserData = (data: Partial<IUser>) => {
    name.value = data.name || name.value;
    username.value = data.username || username.value;
    email.value = data.email || email.value;
    recoveryEmail.value = data.recovery_email || recoveryEmail.value;
    localStorage.setItem("name", name.value);
    localStorage.setItem("user", username.value);
    localStorage.setItem("email", email.value);
    localStorage.setItem("recovery_email", recoveryEmail.value);
  };

  const setShowWelcomeScreen = (tenantID: string) => {
    const current = JSON.parse(localStorage.getItem("namespacesWelcome") || "{}") as Record<string, boolean>;
    current[tenantID] = true;
    localStorage.setItem("namespacesWelcome", JSON.stringify(current));
  };

  return {
    token,
    username,
    name,
    tenantId,
    email,
    id,
    role,
    isAdmin,
    recoveryEmail,
    isMfaEnabled,
    authMethods,
    recoveryCode,
    isRecoveringMfa,
    loginTimeout,
    disableTimeout,
    mfaToken,

    showRecoveryModal,
    isLoggedIn,
    showForceRecoveryMail,

    persistAuth,
    login,
    loginWithToken,
    validateMfa,
    recoverMfa,
    disableMfa,
    generateMfa,
    getUserInfo,
    enterInvitedNamespace,
    requestMfaReset,
    resetMfa,
    enableMfa,
    deleteUser,
    logout,
    updateUserData,
    setShowWelcomeScreen,
  };
});

export default useAuthStore;
