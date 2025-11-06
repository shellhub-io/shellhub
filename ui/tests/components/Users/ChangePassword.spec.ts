import { createPinia, setActivePinia } from "pinia";
import { createVuetify } from "vuetify";
import { DOMWrapper, flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import ChangePassword from "@/components/User/ChangePassword.vue";
import { usersApi } from "@/api/http";
import { SnackbarInjectionKey } from "@/plugins/snackbar";
import useAuthStore from "@/store/modules/auth";
import useUsersStore from "@/store/modules/users";

const mockSnackbar = {
  showSuccess: vi.fn(),
  showError: vi.fn(),
};

type ChangePasswordWrapper = VueWrapper<InstanceType<typeof ChangePassword>>;

describe("Change Password", () => {
  let wrapper: ChangePasswordWrapper;
  setActivePinia(createPinia());
  const authStore = useAuthStore();
  const usersStore = useUsersStore();
  const vuetify = createVuetify();
  const mockUsersApi = new MockAdapter(usersApi.getAxios());

  const authData = {
    token: "",
    username: "test",
    name: "test",
    tenantId: "fake-tenant",
    email: "test@test.com",
    id: "507f1f77bcf86cd799439011",
    role: "owner",
    mfa: {
      enable: false,
      validate: false,
    },
  };

  beforeEach(() => {
    mockUsersApi.onGet("http://localhost:3000/api/auth/user").reply(200, authData);
    mockUsersApi.onGet("http://localhost:3000/api/auth/user").reply(200, authData);

    authStore.$patch(authData);

    wrapper = mount(ChangePassword, {
      global: {
        plugins: [vuetify],
        provide: { [SnackbarInjectionKey]: mockSnackbar },
      },
      props: { modelValue: true },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders components", async () => {
    wrapper.vm.showDialog = true;
    const dialog = new DOMWrapper(document.body);
    await flushPromises();
    expect(dialog.find('[data-test="password-change-dialog"]').exists()).toBe(true);
    expect(dialog.find('[data-test="password-input"]').exists()).toBe(true);
    expect(dialog.find('[data-test="new-password-input"]').exists()).toBe(true);
    expect(dialog.find('[data-test="confirm-new-password-input"]').exists()).toBe(true);
    expect(dialog.find('[data-test="close-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="change-password-btn"]').exists()).toBe(true);
  });

  it("Successfully Change Password", async () => {
    mockUsersApi.onPatch("http://localhost:3000/api/users").reply(200);

    const storeSpy = vi.spyOn(usersStore, "patchPassword");

    wrapper.vm.showDialog = true;
    await flushPromises();

    await wrapper.findComponent('[data-test="password-input"]').setValue("xxxxxx");
    await wrapper.findComponent('[data-test="new-password-input"]').setValue("x1x2x3");
    await wrapper.findComponent('[data-test="confirm-new-password-input"]').setValue("x1x2x3");
    await wrapper.findComponent('[data-test="change-password-btn"]').trigger("click");

    await flushPromises();
    expect(storeSpy).toHaveBeenCalledWith({
      name: "test",
      username: "test",
      email: "test@test.com",
      recovery_email: "",
      currentPassword: "xxxxxx",
      newPassword: "x1x2x3",
    });
  });

  it("Fails to Change Password", async () => {
    mockUsersApi.onPatch("http://localhost:3000/api/users").reply(403);

    const storeSpy = vi.spyOn(usersStore, "patchPassword");

    wrapper.vm.showDialog = true;
    await flushPromises();

    await wrapper.findComponent('[data-test="password-input"]').setValue("xxxxxx");
    await wrapper.findComponent('[data-test="new-password-input"]').setValue("x1x2x3");
    await wrapper.findComponent('[data-test="confirm-new-password-input"]').setValue("x1x2x3");

    await wrapper.findComponent('[data-test="change-password-btn"]').trigger("click");
    await flushPromises();

    expect(storeSpy).toHaveBeenCalledWith({
      name: "test",
      username: "test",
      email: "test@test.com",
      recovery_email: "",
      currentPassword: "xxxxxx",
      newPassword: "x1x2x3",
    });

    expect(mockSnackbar.showError).toHaveBeenCalledWith("An error occurred while updating the password.");
  });
});
