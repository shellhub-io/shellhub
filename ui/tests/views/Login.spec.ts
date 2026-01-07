import { flushPromises, VueWrapper } from "@vue/test-utils";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { RouteLocationAsRelativeGeneric, Router } from "vue-router";
import { mountComponent, mockSnackbar } from "@tests/utils/mount";
import createCleanRouter from "@tests/utils/router";
import { createAxiosError } from "@tests/utils/axiosError";
import Login from "@/views/Login.vue";
import { envVariables } from "@/envVariables";
import useAuthStore from "@/store/modules/auth";
import * as handleErrorModule from "@/utils/handleError";
import { routes } from "@/router";

vi.mock("@/envVariables", () => ({
  envVariables: {
    isCloud: true,
    isEnterprise: false,
  },
}));

const mockRoutes = [
  ...routes,
  // Add Login route without beforeEnter guard
  { path: "/login", name: "Login", meta: { layout: "LoginLayout", requiresAuth: false }, component: Login },
];

describe("Login View", () => {
  let wrapper: VueWrapper<InstanceType<typeof Login>>;
  let router: Router;
  let authStore: ReturnType<typeof useAuthStore>;

  const mountWrapper = async (options: {
    route?: RouteLocationAsRelativeGeneric;
    initialState?: Record<string, object>;
    stubActions?: boolean;
  } = {}) => {
    const {
      route = { name: "Login" },
      initialState = {},
      stubActions = false,
    } = options;

    router = createCleanRouter(mockRoutes);
    await router.push(route);
    await router.isReady();

    wrapper = mountComponent(Login, {
      global: { plugins: [router] },
      piniaOptions: { initialState, stubActions },
    });

    authStore = useAuthStore();
    await flushPromises();
  };

  afterEach(() => {
    vi.clearAllMocks();
    wrapper?.unmount();
  });

  describe("when on cloud environment", () => {
    beforeEach(async () => {
      envVariables.isCloud = true;
      envVariables.isEnterprise = false;
      await mountWrapper();
    });

    it("displays login form fields", () => {
      expect(wrapper.find('[data-test="username-text"] input').exists()).toBe(true);
      expect(wrapper.find('[data-test="password-text"] input').exists()).toBe(true);
      expect(wrapper.find('[data-test="login-btn"]').exists()).toBe(true);
    });

    it("displays forgot password link", () => {
      const forgotPasswordCard = wrapper.find('[data-test="forgotPassword-card"]');
      expect(forgotPasswordCard.text()).toContain("Forgot your Password?");
    });

    it("displays sign up link", () => {
      const signUpCard = wrapper.find('[data-test="isCloud-card"]');
      expect(signUpCard.text()).toContain("Don't have an account?");
      expect(signUpCard.text()).toContain("Sign up here");
    });

    it("does not display SSO button", () => {
      expect(wrapper.find('[data-test="sso-btn"]').exists()).toBe(false);
      expect(wrapper.find('[data-test="or-divider-sso"]').exists()).toBe(false);
    });

    it("enables username and password fields by default", () => {
      const usernameField = wrapper.find('[data-test="username-text"] input');
      const passwordField = wrapper.find('[data-test="password-text"] input');

      expect(usernameField.attributes("disabled")).toBeUndefined();
      expect(passwordField.attributes("disabled")).toBeUndefined();
    });

    it("disables login button when form is invalid", () => {
      const loginButton = wrapper.find('[data-test="login-btn"]');
      expect(loginButton.attributes("disabled")).toBeDefined();
    });

    it("enables login button when form is valid", async () => {
      await wrapper.find('[data-test="username-text"] input').setValue("testuser");
      await wrapper.find('[data-test="password-text"] input').setValue("password");
      await flushPromises();

      const loginButton = wrapper.find('[data-test="login-btn"]');
      expect(loginButton.attributes("disabled")).toBeUndefined();
    });
  });

  describe("when on enterprise environment with SSO", () => {
    beforeEach(async () => {
      wrapper.unmount();
      envVariables.isCloud = false;
      envVariables.isEnterprise = true;
      await mountWrapper({
        initialState: {
          users: {
            systemInfo: {
              setup: true,
              authentication: { saml: true, local: false },
            },
          },
        },
      });
    });

    it("displays SSO login button", () => {
      expect(wrapper.find('[data-test="sso-btn"]').text()).toContain("Login with SSO");
      expect(wrapper.find('[data-test="or-divider-sso"]').exists()).toBe(true);
    });

    it("disables username and password fields when local auth is disabled", () => {
      const usernameField = wrapper.find('[data-test="username-text"] input').element as HTMLInputElement;
      const passwordField = wrapper.find('[data-test="password-text"] input').element as HTMLInputElement;

      expect(usernameField.disabled).toBe(true);
      expect(passwordField.disabled).toBe(true);
    });

    it("disables login button when local auth is disabled", () => {
      const loginButton = wrapper.find('[data-test="login-btn"]');
      expect(loginButton.attributes("disabled")).toBeDefined();
    });
  });

  describe("user login flow", () => {
    beforeEach(async () => {
      envVariables.isCloud = true;
      envVariables.isEnterprise = false;
      await mountWrapper();
    });

    it("calls login action with username and password on form submit", async () => {
      const loginSpy = vi.spyOn(authStore, "login").mockResolvedValueOnce(undefined);
      const routerPushSpy = vi.spyOn(router, "push");

      await wrapper.find('[data-test="username-text"] input').setValue("testuser");
      await wrapper.find('[data-test="password-text"] input').setValue("password123");
      await wrapper.find('[data-test="form"]').trigger("submit");
      await flushPromises();

      expect(loginSpy).toHaveBeenCalledWith({
        username: "testuser",
        password: "password123",
      });

      expect(routerPushSpy).toHaveBeenCalledWith({
        path: "/",
        query: {},
      });
    });

    it("redirects to MFA page when MFA is enabled", async () => {
      vi.spyOn(authStore, "login").mockResolvedValueOnce(undefined);
      authStore.isMfaEnabled = true;
      const routerPushSpy = vi.spyOn(router, "push");

      await wrapper.find('[data-test="username-text"] input').setValue("testuser");
      await wrapper.find('[data-test="password-text"] input').setValue("password123");
      await wrapper.find('[data-test="form"]').trigger("submit");
      await flushPromises();

      expect(routerPushSpy).toHaveBeenCalledWith({ name: "MfaLogin" });
    });

    it("preserves redirect path from query params", async () => {
      await mountWrapper({ route: { name: "Login", query: { redirect: "/devices" } }, stubActions: true });

      vi.spyOn(authStore, "login").mockResolvedValueOnce(undefined);
      const routerPushSpy = vi.spyOn(router, "push");

      await wrapper.find('[data-test="username-text"] input').setValue("testuser");
      await wrapper.find('[data-test="password-text"] input').setValue("password123");
      await wrapper.find('[data-test="form"]').trigger("submit");
      await flushPromises();

      expect(routerPushSpy).toHaveBeenCalledWith({
        path: "/devices",
        query: {},
      });
    });
  });

  describe("error handling", () => {
    beforeEach(async () => {
      envVariables.isCloud = true;
      envVariables.isEnterprise = false;
      await mountWrapper();
    });

    it("displays error message for invalid credentials (401)", async () => {
      const error = createAxiosError(401, "Unauthorized");
      vi.spyOn(authStore, "login").mockRejectedValueOnce(error);

      await wrapper.find('[data-test="username-text"] input').setValue("testuser");
      await wrapper.find('[data-test="password-text"] input').setValue("wrongpassword");
      await wrapper.find('[data-test="form"]').trigger("submit");
      await flushPromises();

      const alert = wrapper.find('[data-test="invalid-login-alert"]');
      expect(alert.exists()).toBe(true);
      expect(alert.text()).toContain("Invalid login credentials");
      expect(alert.text()).toContain("Your password is incorrect or this account doesn't exist");
    });

    it("redirects to confirm account page for unconfirmed user (403)", async () => {
      const error = createAxiosError(403, "Forbidden");
      vi.spyOn(authStore, "login").mockRejectedValueOnce(error);
      const routerPushSpy = vi.spyOn(router, "push").mockImplementationOnce(() => Promise.resolve());

      await wrapper.find('[data-test="username-text"] input').setValue("testuser");
      await wrapper.find('[data-test="password-text"] input').setValue("password123");
      await wrapper.find('[data-test="form"]').trigger("submit");
      await flushPromises();

      expect(routerPushSpy).toHaveBeenCalledWith({
        name: "ConfirmAccount",
        query: { username: "testuser" },
      });
    });

    it("displays lockout message for too many failed attempts (429)", async () => {
      const error = createAxiosError(429, "Too Many Requests");
      authStore.loginTimeout = 300;
      vi.spyOn(authStore, "login").mockRejectedValueOnce(error);

      await wrapper.find('[data-test="username-text"] input').setValue("testuser");
      await wrapper.find('[data-test="password-text"] input').setValue("wrongpassword");
      await wrapper.find('[data-test="form"]').trigger("submit");
      await flushPromises();

      const alert = wrapper.find('[data-test="invalid-login-alert"]');
      expect(alert.exists()).toBe(true);
      expect(alert.text()).toContain("Your account is blocked");
      expect(alert.text()).toContain("There was too many failed login attempts");
    });

    it("displays error snackbar for server errors", async () => {
      const error = createAxiosError(500, "Internal Server Error");
      vi.spyOn(handleErrorModule, "default").mockImplementation(() => { });
      vi.spyOn(authStore, "login").mockRejectedValueOnce(error);

      await wrapper.find('[data-test="username-text"] input').setValue("testuser");
      await wrapper.find('[data-test="password-text"] input').setValue("password123");
      await wrapper.find('[data-test="form"]').trigger("submit");
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Something went wrong in our server. Please try again later.");
    });
  });

  describe("alert messages", () => {
    it("displays alert when redirected from accept invite", async () => {
      await mountWrapper({
        route: { name: "Login", query: { redirect: "/accept-invite/test-id" } },
        initialState: {
          auth: { isLoggedIn: false },
        },
      });

      const alert = wrapper.find('[data-test="user-status-alert"]');
      expect(alert.exists()).toBe(true);
      expect(alert.text()).toContain("Please login before accepting any namespace invitation");
    });

    it("displays alert when user account is not confirmed", async () => {
      await mountWrapper({
        initialState: {
          namespaces: {
            userStatus: "not-confirmed",
          },
        },
      });

      const alert = wrapper.find('[data-test="user-status-alert"]');
      expect(alert.exists()).toBe(true);
      expect(alert.text()).toContain("Your account is not confirmed");
    });

    it("displays alert when SSO has missing assertions", async () => {
      await mountWrapper({ route: { name: "Login", query: { missing_assertions: "true" } } });

      const alert = wrapper.find('[data-test="user-status-alert"]');
      expect(alert.exists()).toBe(true);
      expect(alert.text()).toContain("The SSO configuration is incomplete");
    });
  });
});
