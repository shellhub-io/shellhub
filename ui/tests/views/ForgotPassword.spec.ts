import { setActivePinia, createPinia } from "pinia";
import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import ForgotPassword from "@/views/ForgotPassword.vue";
import { usersApi } from "@/api/http";
import { router } from "@/router";
import { envVariables } from "@/envVariables";
import { SnackbarInjectionKey } from "@/plugins/snackbar";
import useUsersStore from "@/store/modules/users";

type ForgotPasswordWrapper = VueWrapper<InstanceType<typeof ForgotPassword>>;

const mockSnackbar = { showError: vi.fn() };

describe("Forgot Password", () => {
  let wrapper: ForgotPasswordWrapper;
  const vuetify = createVuetify();
  setActivePinia(createPinia());
  const usersStore = useUsersStore();
  const mockUsersApi = new MockAdapter(usersApi.getAxios());

  beforeEach(() => {
    envVariables.isCloud = true;

    wrapper = mount(ForgotPassword, {
      global: {
        plugins: [vuetify, router],
        provide: { [SnackbarInjectionKey]: mockSnackbar },
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders the template with data", () => {
    expect(wrapper.find('[data-test="account-text"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="title-text"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="body-text"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="forgotPassword-btn"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="isCloud-card"]').exists()).toBe(true);
  });

  it("Calls the Forgot Password action when the button is clicked", async () => {
    mockUsersApi.onPost("http://localhost:3000/api/user/recover_password").reply(200);

    const storeSpy = vi.spyOn(usersStore, "recoverPassword");

    await wrapper.findComponent('[data-test="account-text"]').setValue("testuser");
    await wrapper.find('[data-test="forgotPassword-btn"]').trigger("submit");

    await flushPromises();

    expect(storeSpy).toHaveBeenCalledWith("testuser");
  });

  it("Displays success message on successful email submission", async () => {
    mockUsersApi.onPost("http://localhost:3000/api/user/recover_password").reply(200);

    await wrapper.findComponent('[data-test="account-text"]').setValue("testuser");
    await wrapper.find('[data-test="forgotPassword-btn"]').trigger("submit");

    await flushPromises();

    expect(wrapper.find('[data-test="success-text"]').exists()).toBe(true);
  });
});
