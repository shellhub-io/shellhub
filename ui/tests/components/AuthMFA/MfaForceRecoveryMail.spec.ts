import { createPinia, setActivePinia } from "pinia";
import { createVuetify } from "vuetify";
import { DOMWrapper, flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, afterEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import MfaForceRecoveryMail from "@/components/AuthMFA/MfaForceRecoveryMail.vue";
import { usersApi } from "@/api/http";
import { store, key } from "@/store";
import { router } from "@/router";
import { SnackbarPlugin } from "@/plugins/snackbar";
import useAuthStore from "@/store/modules/auth";

type MfaForceRecoveryMailWrapper = VueWrapper<InstanceType<typeof MfaForceRecoveryMail>>;

describe("Force Adding a Recovery Mail", () => {
  let wrapper: MfaForceRecoveryMailWrapper;
  setActivePinia(createPinia());
  const authStore = useAuthStore();
  const vuetify = createVuetify();

  const mockUsersApi = new MockAdapter(usersApi.getAxios());

  const authData = {
    token: "",
    username: "test",
    name: "test",
    tenantId: "fake-tenant-data",
    email: "test@test.com",
    id: "xxxxxxxx",
    recoveryEmail: "recover@mail.com",
    role: "owner",
    mfa: true,
  };

  beforeEach(() => {
    authStore.$patch(authData);

    wrapper = mount(MfaForceRecoveryMail, {
      global: {
        plugins: [[store, key], vuetify, router, SnackbarPlugin],
      },
    });
  });

  afterEach(() => {
    vi.useRealTimers();
    vi.restoreAllMocks();
    mockUsersApi.reset();
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders the components", async () => {
    const dialog = new DOMWrapper(document.body);
    wrapper.vm.showDialog = true;
    await flushPromises();

    expect(dialog.find('[data-test="card-dialog"]').exists()).toBe(true);
    expect(dialog.find('[data-test="dialog-title"]').exists()).toBe(true);
    expect(dialog.find('[data-test="dialog-text"]').exists()).toBe(true);
    expect(dialog.find('[data-test="recovery-email-text"]').exists()).toBe(true);
    expect(dialog.find('[data-test="save-btn"]').exists()).toBe(true);
  });

  it("Adds a recovery mail", async () => {
    wrapper.vm.showDialog = true;
    await flushPromises();
    const storeSpy = vi.spyOn(store, "dispatch");
    mockUsersApi.onPatch("http://localhost:3000/api/users").reply(200);
    await wrapper.findComponent('[data-test="recovery-email-text"]').setValue("test2@test.com");
    await wrapper.findComponent('[data-test="save-btn"]').trigger("click");
    await flushPromises();
    expect(storeSpy).toHaveBeenCalledWith("users/patchData", { id: "xxxxxxxx", recovery_email: "test2@test.com" });
    expect(wrapper.vm.recoveryEmailError).toBe(undefined);
  });

  it("Adds a recovery mail (Fail)", async () => {
    wrapper.vm.showDialog = true;
    await flushPromises();
    const storeSpy = vi.spyOn(store, "dispatch");
    mockUsersApi.onPatch("http://localhost:3000/api/users").reply(400);
    await wrapper.findComponent('[data-test="recovery-email-text"]').setValue("test");
    await wrapper.findComponent('[data-test="save-btn"]').trigger("click");
    await flushPromises();
    expect(storeSpy).toHaveBeenCalledWith("users/patchData", { id: "xxxxxxxx", recovery_email: "test" });
    expect(wrapper.vm.recoveryEmailError).toBe("This recovery email is invalid");
  });

  it("Adds a recovery mail (Fail, Same Email)", async () => {
    wrapper.vm.showDialog = true;
    await flushPromises();
    const storeSpy = vi.spyOn(store, "dispatch");
    mockUsersApi.onPatch("http://localhost:3000/api/users").reply(409);
    await wrapper.findComponent('[data-test="recovery-email-text"]').setValue("test@test.com");
    await wrapper.findComponent('[data-test="save-btn"]').trigger("click");
    await flushPromises();
    expect(storeSpy).toHaveBeenCalledWith("users/patchData", { id: "xxxxxxxx", recovery_email: "test@test.com" });
    expect(wrapper.vm.recoveryEmailError).toBe("This recovery email is already in use");
  });
});
