import { VueWrapper, flushPromises } from "@vue/test-utils";
import { Router } from "vue-router";
import { describe, expect, it, beforeEach, vi, afterEach } from "vitest";
import { mountComponent, mockSnackbar } from "@tests/utils/mount";
import { createCleanRouter } from "@tests/utils/router";
import ValidationAccount from "@/views/ValidationAccount.vue";
import useUsersStore from "@/store/modules/users";
import { createAxiosError } from "@tests/utils/axiosError";
import * as usersApi from "@/store/api/users";

vi.mock("@/store/api/users");

describe("ValidationAccount View", () => {
  let wrapper: VueWrapper<InstanceType<typeof ValidationAccount>>;
  let router: Router;
  let usersStore: ReturnType<typeof useUsersStore>;

  const email = "test@test.com";
  const token = "test-token";

  const mountWrapper = async (mockError?: Error) => {
    router = createCleanRouter();
    await router.push({ name: "ValidationAccount", query: { email, token } });
    await router.isReady();
    vi.spyOn(router, "push").mockResolvedValue();

    if (mockError) vi.mocked(usersApi.validateAccount).mockRejectedValue(mockError);

    wrapper = mountComponent(ValidationAccount, {
      global: { plugins: [router] },
      piniaOptions: { stubActions: !mockError },
    });

    usersStore = useUsersStore();

    await flushPromises();
  };

  afterEach(() => {
    vi.clearAllMocks();
    wrapper?.unmount();
  });

  describe("when page loads", () => {
    beforeEach(() => mountWrapper());

    it("renders the verification title", () => {
      expect(wrapper.find('[data-test="verification-title"]').text()).toContain("Verification Account");
    });

    it("displays back to login link", () => {
      const backToLogin = wrapper.find('[data-test="back-to-login"]');
      expect(backToLogin.exists()).toBe(true);
      expect(backToLogin.text()).toContain("Back to");
      expect(backToLogin.text()).toContain("Login");
    });
  });

  describe("when validation is in progress", () => {
    it("displays processing message initially", async () => {
      wrapper?.unmount();

      router = createCleanRouter();
      await router.push({ name: "ValidationAccount", query: { email, token } });
      await router.isReady();

      wrapper = mountComponent(ValidationAccount, { global: { plugins: [router] } });

      expect(wrapper.find('[data-test="processing-card-text"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="processing-card-text"]').text()).toContain("Processing activation.");
    });
  });

  describe("when validation succeeds", () => {
    beforeEach(() => mountWrapper());

    it("calls validateAccount with correct parameters", () => {
      expect(usersStore.validateAccount).toHaveBeenCalledWith({
        email,
        token,
      });
    });

    it("displays success message", () => {
      expect(wrapper.find('[data-test="success-card-text"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="success-card-text"]').text()).toContain("Congrats and welcome to ShellHub.");
    });

    it("displays success snackbar message", () => {
      expect(mockSnackbar.showSuccess).toHaveBeenCalledWith("Your account has been activated successfully.");
    });

    it("redirects to login after delay", async () => {
      vi.useFakeTimers();
      await mountWrapper();
      const pushSpy = vi.spyOn(router, "push").mockResolvedValue();

      vi.advanceTimersByTime(4000);
      await flushPromises();

      expect(pushSpy).toHaveBeenCalledWith({ path: "/login" });
      vi.useRealTimers();
    });
  });

  describe("when validation fails with 400 error", () => {
    beforeEach(() => mountWrapper(createAxiosError(400, "Bad Request")));

    it("displays failed message", () => {
      expect(wrapper.find('[data-test="failed-card-text"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="failed-card-text"]').text()).toContain(
        "There was a problem activating your account. Go to the login page, login to receive another email with the activation link.",
      );
    });

    it("displays error snackbar message", () => {
      expect(mockSnackbar.showError).toHaveBeenCalledWith("There was a problem activating your account.");
    });

    it("does not redirect to login", async () => {
      vi.useFakeTimers();
      const pushSpy = vi.spyOn(router, "push");

      vi.advanceTimersByTime(4000);
      await flushPromises();

      expect(pushSpy).not.toHaveBeenCalled();
      vi.useRealTimers();
    });
  });

  describe("when validation fails with 404 error", () => {
    beforeEach(() => mountWrapper(createAxiosError(404, "Not Found")));

    it("displays failed token message", () => {
      expect(wrapper.find('[data-test="failed-token-card-text"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="failed-token-card-text"]').text()).toContain(
        "Your account activation token has expired. Go to the login page, login to receive another email with the activation link.",
      );
    });

    it("displays error snackbar message", () => {
      expect(mockSnackbar.showError).toHaveBeenCalledWith("There was a problem activating your account.");
    });

    it("does not redirect to login", async () => {
      vi.useFakeTimers();
      const pushSpy = vi.spyOn(router, "push");

      vi.advanceTimersByTime(4000);
      await flushPromises();

      expect(pushSpy).not.toHaveBeenCalled();
      vi.useRealTimers();
    });
  });
});
