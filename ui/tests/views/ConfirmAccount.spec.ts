import { createPinia, setActivePinia } from "pinia";
import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, afterEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import ConfirmAccount from "@/views/ConfirmAccount.vue";
import { usersApi } from "@/api/http";
import { router } from "@/router";
import { SnackbarInjectionKey } from "@/plugins/snackbar";
import useUsersStore from "@/store/modules/users";

type ConfirmAccountWrapper = VueWrapper<InstanceType<typeof ConfirmAccount>>;
const username = "test";

const mockSnackbar = {
  showError: vi.fn(),
  showSuccess: vi.fn(),
};

describe("Confirm Account", () => {
  let wrapper: ConfirmAccountWrapper;
  const vuetify = createVuetify();
  setActivePinia(createPinia());
  const usersStore = useUsersStore();
  const mockUsersApi = new MockAdapter(usersApi.getAxios());
  beforeEach(async () => {
    await router.push(`/confirm-account?username=${username}`);

    wrapper = mount(ConfirmAccount, {
      global: {
        plugins: [vuetify, router],
        provide: { [SnackbarInjectionKey]: mockSnackbar },
      },
    });
  });

  afterEach(() => {
    wrapper.unmount();
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders the template with data", () => {
    expect(wrapper.find('[data-test="title"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="subtitle"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="resendEmail-btn"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="isCloud-card"]').exists()).toBe(true);
  });

  it("Resends an email to the user", async () => {
    const resendEmailSpy = vi.spyOn(usersStore, "resendEmail");

    mockUsersApi.onPost("http://localhost:3000/api/user/resend_email").reply(200);
    await wrapper.findComponent('[data-test="resendEmail-btn"]').trigger("click");
    await flushPromises();

    expect(resendEmailSpy).toHaveBeenCalledWith(username);
  });

  it("Error case on resends an email to the user", async () => {
    mockUsersApi.onPost("http://localhost:3000/api/user/resend_email").reply(400);
    await wrapper.findComponent('[data-test="resendEmail-btn"]').trigger("click");
    await flushPromises();

    expect(mockSnackbar.showError).toHaveBeenCalledWith("An error occurred while sending the email. Please try again.");
  });
});
