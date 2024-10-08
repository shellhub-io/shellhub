import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, afterEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import AccountCreated from "@/components/Account/AccountCreated.vue";
import { usersApi } from "@/api/http";
import { store, key } from "@/store";
import { router } from "@/router";
import { envVariables } from "@/envVariables";
import { SnackbarPlugin } from "@/plugins/snackbar";

type AccountCreatedWrapper = VueWrapper<InstanceType<typeof AccountCreated>>;

describe("Account Created", () => {
  let wrapper: AccountCreatedWrapper;
  const vuetify = createVuetify();

  let mock: MockAdapter;

  beforeEach(() => {
    vi.useFakeTimers();

    envVariables.isCloud = true;

    mock = new MockAdapter(usersApi.getAxios());
  });

  afterEach(() => {
    vi.useRealTimers();
    vi.restoreAllMocks();
    mock.reset();
  });

  describe("With messageKind = 'normal'", () => {
    beforeEach(() => {
      wrapper = mount(AccountCreated, {
        global: {
          plugins: [[store, key], vuetify, router, SnackbarPlugin],
        },
        props: {
          messageKind: "normal",
          show: true,
          username: "testUser",
        },
      });
    });

    it("Is a Vue instance", () => {
      expect(wrapper.vm).toBeTruthy();
    });

    it("Renders the component", () => {
      expect(wrapper.html()).toMatchSnapshot();
    });

    it("Renders the normal message template", () => {
      expect(wrapper.find('[data-test="accountCreated-card"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="accountCreated-normal-message"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="resendEmail-btn"]').exists()).toBe(true);
      expect(wrapper.text()).toContain("Thank you for registering an account on ShellHub.");
      expect(wrapper.text()).toContain("An email was sent with a confirmation link.");
    });

    it("Resends email", async () => {
      const storeSpy = vi.spyOn(store, "dispatch");

      mock.onPost("http://localhost:3000/api/user/resend_email").reply(200);

      await wrapper.find('[data-test="resendEmail-btn"]').trigger("click");

      await flushPromises();

      expect(storeSpy).toHaveBeenCalledWith("users/resendEmail", "testUser");
      expect(storeSpy).toHaveBeenCalledWith("snackbar/showSnackbarSuccessAction", "resend email");
    });
  });

  describe("With messageKind = 'sig'", () => {
    beforeEach(() => {
      wrapper = mount(AccountCreated, {
        global: {
          plugins: [[store, key], vuetify, router, SnackbarPlugin],
        },
        props: {
          messageKind: "sig",
          show: true,
          username: "testUser",
        },
      });
    });

    it("Is a Vue instance", () => {
      expect(wrapper.vm).toBeTruthy();
    });

    it("Renders the sig message template", () => {
      expect(wrapper.find('[data-test="accountCreated-card"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="accountCreated-sig-message"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="redirect-btn"]').exists()).toBe(true);
      expect(wrapper.text()).toContain("You will be redirected in 5 seconds");
    });

    it("Handles timeout for redirect", async () => {
      const redirectSpy = vi.spyOn(router, "push");

      await wrapper.find('[data-test="redirect-btn"]').trigger("click");

      await flushPromises();

      expect(redirectSpy).toHaveBeenCalledOnce();
    });
  });
});
