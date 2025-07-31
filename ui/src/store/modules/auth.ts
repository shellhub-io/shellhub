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

const useAuthStore = defineStore("auth", () => {
  const token = ref(localStorage.getItem("token") || "");
  const username = ref(localStorage.getItem("user") || "");
  const name = ref(localStorage.getItem("name") || "");
  const tenantId = ref(localStorage.getItem("tenant") || "");
  const email = ref(localStorage.getItem("email") || "");
  const id = ref(localStorage.getItem("id") || "");
  const role = ref(localStorage.getItem("role") || "");
  const recoveryEmail = ref("");
  const secret = ref("");
  const mfaQrCode = ref("");
  const isMfaEnabled = ref(false);
  const authMethods = ref(["local"]);
  const recoveryCode = ref("");
  const recoveryCodes = ref<number[]>([]);
  const isRecoveringMfa = ref(false);
  const loginTimeout = ref(0);
  const disableTimeout = ref(0);
  const mfaToken = ref<string | undefined>();

  const isLoggedIn = computed(() => !!token.value);
  const showForceRecoveryMail = computed(() => !recoveryEmail.value && isMfaEnabled.value);
  const showRecoveryModal = computed(() => isRecoveringMfa.value && isMfaEnabled.value);

  function persistAuth(data) {
    token.value = data.token || "";
    username.value = data.user || data.username || "";
    name.value = data.name || "";
    tenantId.value = data.tenantId || data.tenant || "";
    email.value = data.email || "";
    id.value = data.id || "";
    role.value = data.role || "";
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
    localStorage.setItem("recovery_email", recoveryEmail.value);
    localStorage.setItem("mfa", String(isMfaEnabled.value));
    localStorage.setItem("namespacesWelcome", JSON.stringify({}));

    apiConfiguration.accessToken = token.value;
    reloadApiConfiguration();
  }

  async function login(userInput: IUserLogin) {
    try {
      const resp = await authApi.login(userInput);
      persistAuth(resp.data);
    } catch (error) {
      const axiosError = error as AxiosError;
      const token = axiosError.response?.headers["x-mfa-token"];
      if (token) {
        isMfaEnabled.value = true;
        localStorage.setItem("mfa", "true");
        mfaToken.value = token;
        return;
      }

      loginTimeout.value = axiosError?.response?.headers["x-account-lockout"];
      throw error;
    }
  }

  async function loginWithToken(tokenInput: string) {
    apiConfiguration.accessToken = tokenInput;
    reloadApiConfiguration();
    const resp = await authApi.getUserInfo();
    persistAuth(resp.data);
  }

  async function validateMfa(verificationCode: string) {
    const resp = await authApi.validateMFA({ token: mfaToken.value as string, code: verificationCode });
    if (resp.status === 200) {
      persistAuth(resp.data);
      mfaToken.value = resp.data.token;
      isMfaEnabled.value = true;
    }
  }

  async function recoverMfa(recoveryCode: string) {
    const resp = await authApi.recoverMfa({ identifier: name.value, recovery_code: recoveryCode });
    if (resp.status === 200) {
      persistAuth(resp.data);
      mfaToken.value = resp.data.token;
      isRecoveringMfa.value = true;
      disableTimeout.value = resp.headers["x-expires-at"];
    }
  }

  async function disableMfa(data: Partial<IMfaDisable>) {
    await authApi.disableMfa(data);
    isMfaEnabled.value = false;
    localStorage.setItem("mfa", "false");
  }

  async function generateMfa() {
    const { data } = await authApi.generateMfa();
    return data as IMfaGenerate;
  }

  async function getUserInfo() {
    const resp = await authApi.getUserInfo();
    if (resp.status === 200) persistAuth(resp.data);
  }

  async function enterInvitedNamespace(tenantId: string) {
    localStorage.removeItem("role");
    const res = await namespaceApi.switchNamespace(tenantId);
    if (res.status === 200) {
      persistAuth({ ...res.data, tenantId });
    }
  }

  async function requestMfaReset() {
    await authApi.requestResetMfa(email.value);
  }

  async function resetMfa(data: IMfaReset) {
    const resp = await authApi.resetMfa(data);
    if (resp.status === 200) persistAuth(resp.data);
  }

  async function enableMfa(data: IMfaEnable) {
    const resp = await authApi.enableMFA(data);
    if (resp.status === 200) isMfaEnabled.value = true;
  }

  function logout() {
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
  }

  async function deleteUser() {
    await authApi.deleteUser();
    logout();
  }

  function updateUserData(data: Partial<IUser>) {
    name.value = data.name || name.value;
    username.value = data.username || username.value;
    email.value = data.email || email.value;
    recoveryEmail.value = data.recovery_email || recoveryEmail.value;
    localStorage.setItem("name", name.value);
    localStorage.setItem("user", username.value);
    localStorage.setItem("email", email.value);
    localStorage.setItem("recovery_email", recoveryEmail.value);
  }

  function setShowWelcomeScreen(tenantID: string) {
    const current = JSON.parse(localStorage.getItem("namespacesWelcome") || "{}");
    current[tenantID] = true;
    localStorage.setItem("namespacesWelcome", JSON.stringify(current));
  }

  return {
    token,
    username,
    name,
    tenantId,
    email,
    id,
    role,
    recoveryEmail,
    secret,
    mfaQrCode,
    isMfaEnabled,
    authMethods,
    recoveryCode,
    recoveryCodes,
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
