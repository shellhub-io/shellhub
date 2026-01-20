import { describe, expect, it, beforeEach, vi, afterEach } from "vitest";
import { Router } from "vue-router";
import { VueWrapper, flushPromises } from "@vue/test-utils";
import { mountComponent, mockSnackbar } from "@tests/utils/mount";
import { createCleanRouter } from "@tests/utils/router";
import { createAxiosError } from "@tests/utils/axiosError";
import useAuthStore from "@/store/modules/auth";
import useUsersStore from "@/store/modules/users";
import AccountCreated from "@/components/Account/AccountCreated.vue";
import SignUp from "@/views/SignUp.vue";
import { routes } from "@/router";

const mockRoutes = [
  ...routes,
  { name: "SignUp", path: "/sign-up", component: SignUp },
];

describe("AccountCreated", () => {
  let wrapper: VueWrapper<InstanceType<typeof AccountCreated>>;
  let router: Router;
  let authStore: ReturnType<typeof useAuthStore>;
  let usersStore: ReturnType<typeof useUsersStore>;
  const mockUsername = "testuser";

  afterEach(() => {
    vi.clearAllMocks();
    vi.useRealTimers();
    wrapper?.unmount();
  });

  describe("normal message mode", () => {
    const mountWrapper = () => {
      router = createCleanRouter(mockRoutes);

      wrapper = mountComponent(AccountCreated, {
        global: { plugins: [router] },
        props: {
          messageKind: "normal",
          show: true,
          username: mockUsername,
        },
      });

      authStore = useAuthStore();
      usersStore = useUsersStore();
    };

    describe("rendering", () => {
      beforeEach(() => mountWrapper());

      it("renders the card when show prop is true", () => {
        const card = wrapper.find('[data-test="account-created-card"]');
        expect(card.exists()).toBe(true);
      });

      it("displays the success title", () => {
        expect(wrapper.text()).toContain("Account Creation Successful");
      });

      it("displays the thank you message", () => {
        const message = wrapper.find('[data-test="account-created-message"]');
        expect(message.text()).toContain("Thank you for registering an account on ShellHub");
      });

      it("displays the normal message with email confirmation text", () => {
        const normalMessage = wrapper.find('[data-test="account-created-normal-message"]');
        expect(normalMessage.exists()).toBe(true);
        expect(normalMessage.text()).toContain("An email was sent with a confirmation link");
        expect(normalMessage.text()).toContain("You need to click on the link to activate your account");
      });

      it("does not display the sig message", () => {
        const sigMessage = wrapper.find('[data-test="account-created-sig-message"]');
        expect(sigMessage.exists()).toBe(false);
      });

      it("displays the email info text", () => {
        const emailInfo = wrapper.find('[data-test="account-created-email-info"]');
        expect(emailInfo.exists()).toBe(true);
        expect(emailInfo.text()).toContain("If you haven't received the email, click on the button");
      });

      it("displays the resend email button", () => {
        const resendBtn = wrapper.find('[data-test="resend-email-btn"]');
        expect(resendBtn.exists()).toBe(true);
        expect(resendBtn.text()).toBe("RESEND EMAIL");
      });
    });

    describe("resending email", () => {
      beforeEach(() => mountWrapper());

      it("calls store action when clicking resend button", async () => {
        const resendBtn = wrapper.find('[data-test="resend-email-btn"]');
        await resendBtn.trigger("click");
        await flushPromises();

        expect(usersStore.resendEmail).toHaveBeenCalledWith(mockUsername);
      });

      it("shows success message after successfully resending email", async () => {
        const resendBtn = wrapper.find('[data-test="resend-email-btn"]');
        await resendBtn.trigger("click");
        await flushPromises();

        expect(mockSnackbar.showSuccess).toHaveBeenCalledWith("Email successfully sent.");
      });

      it("shows error message when resending email fails", async () => {
        vi.mocked(usersStore.resendEmail).mockRejectedValueOnce(
          createAxiosError(500, "Internal Server Error"),
        );

        const resendBtn = wrapper.find('[data-test="resend-email-btn"]');
        await resendBtn.trigger("click");
        await flushPromises();

        expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to send email.");
      });
    });

    describe("visibility control", () => {
      beforeEach(() => mountWrapper());

      it("does not render card when show prop is false", async () => {
        await wrapper.setProps({ show: false });
        await flushPromises();

        const card = wrapper.find('[data-test="account-created-card"]');
        expect(card.exists()).toBe(false);
      });
    });
  });

  describe("sig message mode", () => {
    const mockSignUpToken = "mock-signup-token-123";
    const mockQueryParams = {
      "tenant-id": "fake-tenant",
      "user-id": "fake-id",
      sig: "fake-sig",
    };

    const mountWrapper = async (showProp = true) => {
      router = createCleanRouter(mockRoutes);
      await router.push({ name: "SignUp", query: mockQueryParams });
      await router.isReady();

      wrapper = mountComponent(AccountCreated, {
        global: { plugins: [router] },
        props: {
          messageKind: "sig",
          show: showProp,
          username: mockUsername,
        },
        piniaOptions: {
          initialState: { users: { signUpToken: mockSignUpToken } },
        },
      });

      authStore = useAuthStore();
      usersStore = useUsersStore();
    };

    describe("rendering", () => {
      beforeEach(() => mountWrapper());

      it("renders the card when show prop is true", () => {
        const card = wrapper.find('[data-test="account-created-card"]');
        expect(card.exists()).toBe(true);
      });

      it("displays the sig message with redirect text", () => {
        const sigMessage = wrapper.find('[data-test="account-created-sig-message"]');
        expect(sigMessage.exists()).toBe(true);
        expect(sigMessage.text()).toContain("You will be redirected in 5 seconds");
        expect(sigMessage.text()).toContain("if you weren't redirected, please click the button below");
      });

      it("does not display the normal message", () => {
        const normalMessage = wrapper.find('[data-test="account-created-normal-message"]');
        expect(normalMessage.exists()).toBe(false);
      });

      it("does not display the email info text", () => {
        const emailInfo = wrapper.find('[data-test="account-created-email-info"]');
        expect(emailInfo.exists()).toBe(false);
      });

      it("displays the redirect button", () => {
        const redirectBtn = wrapper.find('[data-test="redirect-btn"]');
        expect(redirectBtn.exists()).toBe(true);
        expect(redirectBtn.text()).toBe("REDIRECT");
      });
    });

    describe("automatic redirect", () => {
      beforeEach(async () => {
        vi.useFakeTimers();
        await mountWrapper(false);
        await wrapper.setProps({ show: true }); // Trigger watch
        await flushPromises();
      });

      it("calls loginWithToken when component becomes visible in sig mode", () => {
        expect(authStore.loginWithToken).toHaveBeenCalledWith(mockSignUpToken);
      });

      it("automatically redirects after 5 seconds", async () => {
        const pushSpy = vi.spyOn(router, "push");

        // Fast-forward 5 seconds
        vi.advanceTimersByTime(5000);
        await flushPromises();

        expect(pushSpy).toHaveBeenCalledWith({
          name: "AcceptInvite",
          query: mockQueryParams,
        });
      });
    });

    describe("manual redirect", () => {
      beforeEach(() => mountWrapper());

      it("redirects when clicking the redirect button", async () => {
        const pushSpy = vi.spyOn(router, "push");
        const redirectBtn = wrapper.find('[data-test="redirect-btn"]');

        await redirectBtn.trigger("click");
        await flushPromises();

        expect(pushSpy).toHaveBeenCalledWith({
          name: "AcceptInvite",
          query: mockQueryParams,
        });
      });

      it("removes redirect parameter from query when redirecting", async () => {
        wrapper.unmount();
        await router.push({ name: "SignUp", query: { ...mockQueryParams, redirect: "/some-path" } });
        await router.isReady();

        wrapper = mountComponent(AccountCreated, {
          global: { plugins: [router] },
          props: {
            messageKind: "sig",
            show: true,
            username: mockUsername,
          },
          piniaOptions: {
            initialState: { users: { signUpToken: mockSignUpToken } },
          },
        });

        const pushSpy = vi.spyOn(router, "push");
        const redirectBtn = wrapper.find('[data-test="redirect-btn"]');

        await redirectBtn.trigger("click");
        await flushPromises();

        expect(pushSpy).toHaveBeenCalledWith({
          name: "AcceptInvite",
          query: mockQueryParams, // Should not include "redirect"
        });
      });
    });

    describe("error handling", () => {
      beforeEach(() => mountWrapper());

      it("handles errors gracefully when redirect fails", async () => {
        const pushSpy = vi.spyOn(router, "push").mockRejectedValueOnce(new Error("Navigation failed"));
        const redirectBtn = wrapper.find('[data-test="redirect-btn"]');

        await redirectBtn.trigger("click");
        await flushPromises();

        expect(pushSpy).toHaveBeenCalled();
        // Component should not crash
        expect(wrapper.exists()).toBe(true);
      });
    });
  });
});
