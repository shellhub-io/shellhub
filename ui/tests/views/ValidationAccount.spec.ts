import { createPinia, setActivePinia } from "pinia";
import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { afterEach, beforeEach, describe, expect, it } from "vitest";
import MockAdapter from "axios-mock-adapter";
import ValidationAccount from "@/views/ValidationAccount.vue";
import { usersApi } from "@/api/http";
import { SnackbarPlugin } from "@/plugins/snackbar";
import { router } from "@/router";

type ValidationAccountWrapper = VueWrapper<InstanceType<typeof ValidationAccount>>;

describe("Validation Account", () => {
  let wrapper: ValidationAccountWrapper;
  setActivePinia(createPinia());
  const vuetify = createVuetify();
  const mockUsersApi = new MockAdapter(usersApi.getAxios());

  beforeEach(async () => {
    await router.push("/validation-account?email=test@test.com&token=test-token");
    mockUsersApi.onGet("http://localhost:3000/api/user/validation_account?email=test%40test.com&token=test-token").reply(200);

    wrapper = mount(ValidationAccount, {
      global: {
        plugins: [vuetify, router, SnackbarPlugin],
      },
    });
  });

  afterEach(() => {
    wrapper.unmount();
  });

  it("Is a Vue instance", async () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", async () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders the template with data", async () => {
    expect(wrapper.find('[data-test="verification-title"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="processing-cardText"]').exists()).toBe(true);
  });

  it("Renders success message", async () => {
    await flushPromises();
    expect(wrapper.find('[data-test="success-cardText"]').exists()).toBe(true);
  });

  it("Redirects to login page when login link is clicked", async () => {
    await flushPromises();
    await wrapper.find('[data-test="login-btn"]').trigger("click");
    expect(router.currentRoute.value.path).toBe("/validation-account");
  });
});
