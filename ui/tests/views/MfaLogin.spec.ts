import { flushPromises, VueWrapper } from "@vue/test-utils";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { RouteLocationAsRelativeGeneric, Router } from "vue-router";
import { mountComponent } from "@tests/utils/mount";
import { createCleanRouter } from "@tests/utils/router";
import { createAxiosError } from "@tests/utils/axiosError";
import MfaLogin from "@/views/MfaLogin.vue";
import useAuthStore from "@/store/modules/auth";
import { routes } from "@/router";

vi.mock("@/store/api/auth");

type MfaLoginWrapper = VueWrapper<InstanceType<typeof MfaLogin>>;

const mockRoutes = [
  ...routes,
  // Add MfaLogin route without beforeEnter guard
  { path: "/mfa-login", name: "MfaLogin", meta: { layout: "LoginLayout", requiresAuth: false }, component: MfaLogin },
];

describe("MfaLogin View", () => {
  let wrapper: MfaLoginWrapper;
  let router: Router;
  let authStore: ReturnType<typeof useAuthStore>;

  const mountWrapper = async (options: {
    route?: RouteLocationAsRelativeGeneric;
    initialState?: Record<string, object>;
    stubActions?: boolean;
    mockError?: Error;
  } = {}) => {
    const {
      route = { name: "MfaLogin" },
      initialState = {},
      mockError,
    } = options;

    router = createCleanRouter(mockRoutes);
    await router.push(route);
    await router.isReady();

    wrapper = mountComponent(MfaLogin, {
      global: { plugins: [router] },
      piniaOptions: { initialState, stubActions: !mockError },
    });

    authStore = useAuthStore();
    if (mockError) vi.mocked(authStore.validateMfa).mockRejectedValueOnce(mockError);
  };

  afterEach(() => {
    vi.clearAllMocks();
    wrapper?.unmount();
  });

  describe("form display", () => {
    beforeEach(() => mountWrapper());

    it("displays MFA form elements", () => {
      expect(wrapper.find('[data-test="title"]').text()).toContain("Multi-factor Authentication");
      expect(wrapper.find('[data-test="sub-title"]').text()).toContain("Verify your identity");
      expect(wrapper.find('[data-test="verification-code"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="verify-btn"]').exists()).toBe(true);
    });

    it("displays recovery code link", () => {
      expect(wrapper.find('[data-test="redirect-recover"]').text()).toContain("Lost your TOTP password?");
    });

    it("disables verify button when verification code is empty", () => {
      const verifyButton = wrapper.find('[data-test="verify-btn"]');
      expect(verifyButton.attributes("disabled")).toBeDefined();
    });

    it("enables verify button when verification code is provided", async () => {
      await wrapper.findComponent('[data-test="verification-code"]').setValue("000000");
      await flushPromises();

      const verifyButton = wrapper.find('[data-test="verify-btn"]');
      expect(verifyButton.attributes("disabled")).toBeUndefined();
    });
  });

  describe("MFA verification flow", () => {
    beforeEach(() => mountWrapper());

    it("calls validateMfa action with verification code on form submit", async () => {
      const routerPushSpy = vi.spyOn(router, "push");

      await wrapper.findComponent('[data-test="verification-code"]').setValue("123456");
      await wrapper.find('[data-test="verify-btn"]').trigger("click");
      await flushPromises();

      expect(authStore.validateMfa).toHaveBeenCalledWith("123456");
      expect(routerPushSpy).toHaveBeenCalledWith("/");
    });

    it("redirects to home page on successful verification", async () => {
      const routerPushSpy = vi.spyOn(router, "push");

      await wrapper.findComponent('[data-test="verification-code"]').setValue("123456");
      await wrapper.find('[data-test="verify-btn"]').trigger("click");
      await flushPromises();

      expect(routerPushSpy).toHaveBeenCalledWith("/");
    });

    it("redirects to recovery page when clicking recovery link", async () => {
      const routerPushSpy = vi.spyOn(router, "push");

      await wrapper.find('[data-test="redirect-recover"]').trigger("click");
      await flushPromises();

      expect(routerPushSpy).toHaveBeenCalledWith({ name: "RecoverMfa" });
    });
  });

  describe("error handling", () => {
    it("displays error alert for invalid verification code (500)", async () => {
      const error = createAxiosError(500, "Internal Server Error");
      await mountWrapper({ mockError: error });

      await wrapper.findComponent('[data-test="verification-code"]').setValue("000000");
      await wrapper.find('[data-test="verify-btn"]').trigger("click");
      await flushPromises();

      const alert = wrapper.find('[data-test="alert-message"]');
      expect(alert.exists()).toBe(true);
      expect(alert.text()).toContain("The verification code sent in your MFA verification is invalid");
    });

    it("displays generic error alert for other errors", async () => {
      const error = createAxiosError(400, "Bad Request");
      await mountWrapper({ mockError: error });

      await wrapper.findComponent('[data-test="verification-code"]').setValue("000000");
      await wrapper.find('[data-test="verify-btn"]').trigger("click");
      await flushPromises();

      const alert = wrapper.find('[data-test="alert-message"]');
      expect(alert.exists()).toBe(true);
      expect(alert.text()).toContain("An error occurred during your MFA verification");
    });

    it("does not navigate on verification failure", async () => {
      const error = createAxiosError(500, "Internal Server Error");
      await mountWrapper({ mockError: error });

      const routerPushSpy = vi.spyOn(router, "push");

      await wrapper.findComponent('[data-test="verification-code"]').setValue("000000");
      await wrapper.find('[data-test="verify-btn"]').trigger("click");
      await flushPromises();

      expect(routerPushSpy).not.toHaveBeenCalled();
    });
  });
});
