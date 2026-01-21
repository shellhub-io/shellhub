import { VueWrapper, flushPromises } from "@vue/test-utils";
import { describe, expect, it, beforeEach, vi, afterEach } from "vitest";
import { mountComponent } from "@tests/utils/mount";
import { createCleanRouter } from "@tests/utils/router";
import SignUp from "@/views/SignUp.vue";
import useUsersStore from "@/store/modules/users";
import { createAxiosError } from "@tests/utils/axiosError";
import { routes } from "@/router";

vi.mock("@/store/api/users");

describe("SignUp View", () => {
  let wrapper: VueWrapper<InstanceType<typeof SignUp>>;
  let usersStore: ReturnType<typeof useUsersStore>;

  const mockRoutes = [
    ...routes,
    { name: "SignUp", path: "/sign-up", component: SignUp },
  ];

  const mountWrapper = async (mockError?: Error) => {
    const router = createCleanRouter(mockRoutes);
    await router.push({ name: "SignUp" });
    await router.isReady();

    wrapper = mountComponent(SignUp, {
      global: { plugins: [router] },
      piniaOptions: { stubActions: !mockError },
    });

    usersStore = useUsersStore();

    if (mockError) vi.mocked(usersStore.signUp).mockRejectedValueOnce(mockError);

    await flushPromises();
  };

  afterEach(() => {
    vi.clearAllMocks();
    wrapper?.unmount();
  });

  describe("when page loads", () => {
    beforeEach(() => mountWrapper());

    it("renders all form fields", () => {
      expect(wrapper.find('[data-test="name-text"]').text()).toContain("Name");
      expect(wrapper.find('[data-test="username-text"]').text()).toContain("Username");
      expect(wrapper.find('[data-test="email-text"]').text()).toContain("Email");
      expect(wrapper.find('[data-test="password-text"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="password-confirm-text"]').exists()).toBe(true);
    });

    it("displays privacy policy and marketing checkboxes", () => {
      expect(wrapper.find('[data-test="accept-privacy-policy-checkbox"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="accept-news-checkbox"]').exists()).toBe(true);
    });

    it("displays create account and login buttons", () => {
      expect(wrapper.find('[data-test="create-account-btn"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="login-btn"]').exists()).toBe(true);
    });

    it("does not show account created message initially", () => {
      expect(wrapper.find('[data-test="accountCreated-component"]').exists()).toBe(false);
    });

    it("does not show privacy policy error initially", () => {
      expect(wrapper.find('[data-test="privacy-policy-error"]').exists()).toBe(false);
    });
  });

  describe("when account creation succeeds", () => {
    beforeEach(() => mountWrapper());

    it("calls signUp with correct parameters", async () => {
      await wrapper.find('[data-test="name-text"] input').setValue("test");
      await wrapper.find('[data-test="username-text"] input').setValue("test");
      await wrapper.find('[data-test="email-text"] input').setValue("test@test.com");
      await wrapper.find('[data-test="password-text"] input').setValue("test123");
      await wrapper.find('[data-test="password-confirm-text"] input').setValue("test123");
      await wrapper.find('[data-test="accept-privacy-policy-checkbox"] input').setValue(true);
      await wrapper.find('[data-test="accept-news-checkbox"] input').setValue(true);
      await wrapper.find('[data-test="create-account-btn"]').trigger("submit");
      await flushPromises();

      expect(usersStore.signUp).toHaveBeenCalledWith({
        name: "test",
        email: "test@test.com",
        username: "test",
        password: "test123",
        confirmPassword: "test123",
        emailMarketing: true,
      });
    });

    it("calls signUp with emailMarketing false when unchecked", async () => {
      await wrapper.find('[data-test="name-text"] input').setValue("test");
      await wrapper.find('[data-test="username-text"] input').setValue("test");
      await wrapper.find('[data-test="email-text"] input').setValue("test@test.com");
      await wrapper.find('[data-test="password-text"] input').setValue("test123");
      await wrapper.find('[data-test="password-confirm-text"] input').setValue("test123");
      await wrapper.find('[data-test="accept-privacy-policy-checkbox"] input').setValue(true);
      await wrapper.find('[data-test="create-account-btn"]').trigger("submit");
      await flushPromises();

      expect(usersStore.signUp).toHaveBeenCalledWith({
        name: "test",
        email: "test@test.com",
        username: "test",
        password: "test123",
        confirmPassword: "test123",
        emailMarketing: false,
      });
    });
  });

  describe("when account creation fails", () => {
    it("handles 400 error with username conflict", async () => {
      await mountWrapper(createAxiosError(400, "Bad Request", ["username"]));

      await wrapper.find('[data-test="name-text"] input').setValue("test");
      await wrapper.find('[data-test="username-text"] input').setValue("test");
      await wrapper.find('[data-test="email-text"] input').setValue("test@test.com");
      await wrapper.find('[data-test="password-text"] input').setValue("test123");
      await wrapper.find('[data-test="password-confirm-text"] input').setValue("test123");
      await wrapper.find('[data-test="accept-privacy-policy-checkbox"] input').setValue(true);
      await wrapper.find('[data-test="create-account-btn"]').trigger("submit");
      await flushPromises();

      expect(usersStore.signUp).toHaveBeenCalled();
      expect(wrapper.find('[data-test="username-text"]').text()).toContain("This username already exists");
    });

    it("handles 400 error with email validation failure", async () => {
      await mountWrapper(createAxiosError(400, "Bad Request", ["email"]));

      await wrapper.find('[data-test="name-text"] input').setValue("test");
      await wrapper.find('[data-test="username-text"] input').setValue("test");
      await wrapper.find('[data-test="email-text"] input').setValue("test@test.com");
      await wrapper.find('[data-test="password-text"] input').setValue("test123");
      await wrapper.find('[data-test="password-confirm-text"] input').setValue("test123");
      await wrapper.find('[data-test="accept-privacy-policy-checkbox"] input').setValue(true);
      await wrapper.find('[data-test="create-account-btn"]').trigger("submit");
      await flushPromises();

      expect(usersStore.signUp).toHaveBeenCalled();
      expect(wrapper.find('[data-test="email-text"]').text()).toContain("This email is invalid!");
    });

    it("handles 400 error with password validation failure", async () => {
      await mountWrapper(createAxiosError(400, "Bad Request", ["password"]));

      await wrapper.find('[data-test="name-text"] input').setValue("test");
      await wrapper.find('[data-test="username-text"] input').setValue("test");
      await wrapper.find('[data-test="email-text"] input').setValue("test@test.com");
      await wrapper.find('[data-test="password-text"] input').setValue("test123");
      await wrapper.find('[data-test="password-confirm-text"] input').setValue("test123");
      await wrapper.find('[data-test="accept-privacy-policy-checkbox"] input').setValue(true);
      await wrapper.find('[data-test="create-account-btn"]').trigger("submit");
      await flushPromises();

      expect(usersStore.signUp).toHaveBeenCalled();
      expect(wrapper.find('[data-test="password-text"]').text()).toContain("This password is invalid!");
    });

    it("handles 400 error with name validation failure", async () => {
      await mountWrapper(createAxiosError(400, "Bad Request", ["name"]));

      await wrapper.find('[data-test="name-text"] input').setValue("test");
      await wrapper.find('[data-test="username-text"] input').setValue("test");
      await wrapper.find('[data-test="email-text"] input').setValue("test@test.com");
      await wrapper.find('[data-test="password-text"] input').setValue("test123");
      await wrapper.find('[data-test="password-confirm-text"] input').setValue("test123");
      await wrapper.find('[data-test="accept-privacy-policy-checkbox"] input').setValue(true);
      await wrapper.find('[data-test="create-account-btn"]').trigger("submit");
      await flushPromises();

      expect(usersStore.signUp).toHaveBeenCalled();
      expect(wrapper.find('[data-test="name-text"]').text()).toContain("This name is invalid!");
    });
  });

  describe("form validation", () => {
    beforeEach(() => mountWrapper());

    it("disables submit button when privacy policy is not accepted", async () => {
      await wrapper.find('[data-test="name-text"] input').setValue("test");
      await wrapper.find('[data-test="username-text"] input').setValue("test");
      await wrapper.find('[data-test="email-text"] input').setValue("test@test.com");
      await wrapper.find('[data-test="password-text"] input').setValue("test123");
      await wrapper.find('[data-test="password-confirm-text"] input').setValue("test123");
      await flushPromises();

      expect(wrapper.find('[data-test="create-account-btn"]').attributes("disabled")).toBeDefined();
    });

    it("enables submit button when all required fields are filled", async () => {
      await wrapper.find('[data-test="name-text"] input').setValue("test");
      await wrapper.find('[data-test="username-text"] input').setValue("test");
      await wrapper.find('[data-test="email-text"] input').setValue("test@test.com");
      await wrapper.find('[data-test="password-text"] input').setValue("test123");
      await wrapper.find('[data-test="password-confirm-text"] input').setValue("test123");
      await wrapper.find('[data-test="accept-privacy-policy-checkbox"] input').setValue(true);
      await flushPromises();

      expect(wrapper.find('[data-test="create-account-btn"]').attributes("disabled")).toBeUndefined();
    });

    it("shows error when passwords do not match", async () => {
      await wrapper.find('[data-test="name-text"] input').setValue("test");
      await wrapper.find('[data-test="username-text"] input').setValue("test");
      await wrapper.find('[data-test="email-text"] input').setValue("test@test.com");
      await wrapper.find('[data-test="password-text"] input').setValue("test123");
      await wrapper.find('[data-test="password-confirm-text"] input').setValue("different");
      await wrapper.find('[data-test="accept-privacy-policy-checkbox"] input').setValue(true);
      await flushPromises();

      const confirmPasswordField = wrapper.find('[data-test="password-confirm-text"]');
      expect(confirmPasswordField.text()).toContain("Passwords do not match");
      expect(wrapper.find('[data-test="create-account-btn"]').attributes("disabled")).toBeDefined();
    });
  });
});
