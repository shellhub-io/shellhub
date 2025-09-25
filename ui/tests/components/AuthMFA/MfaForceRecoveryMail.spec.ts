import { createPinia, setActivePinia } from "pinia";
import { createVuetify } from "vuetify";
import { DOMWrapper, flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import MfaForceRecoveryMail from "@/components/AuthMFA/MfaForceRecoveryMail.vue";
import { usersApi } from "@/api/http";
import { router } from "@/router";
import { SnackbarPlugin } from "@/plugins/snackbar";
import useAuthStore from "@/store/modules/auth";
import useUsersStore from "@/store/modules/users";

type MfaForceRecoveryMailWrapper = VueWrapper<InstanceType<typeof MfaForceRecoveryMail>>;

describe("Force Adding a Recovery Mail", () => {
  let wrapper: MfaForceRecoveryMailWrapper;
  setActivePinia(createPinia());
  const authStore = useAuthStore();
  const usersStore = useUsersStore();
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
        plugins: [vuetify, router, SnackbarPlugin],
      },
      props: { modelValue: true },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", async () => {
    await flushPromises();
    const dialog = new DOMWrapper(document.body);
    expect(dialog.html()).toMatchSnapshot();
  });

  it("Adds a recovery mail", async () => {
    await flushPromises();
    const storeSpy = vi.spyOn(usersStore, "patchData");
    mockUsersApi.onPatch("http://localhost:3000/api/users").reply(200);
    await wrapper.findComponent('[data-test="recovery-email-text"]').setValue("test2@test.com");
    await wrapper.findComponent('[data-test="save-btn"]').trigger("click");
    await flushPromises();
    expect(storeSpy).toHaveBeenCalledWith({ recovery_email: "test2@test.com" });
    expect(wrapper.vm.recoveryEmailError).toBe(undefined);
  });

  it("Adds a recovery mail (Fail)", async () => {
    await flushPromises();
    const storeSpy = vi.spyOn(usersStore, "patchData");
    mockUsersApi.onPatch("http://localhost:3000/api/users").reply(400);
    await wrapper.findComponent('[data-test="recovery-email-text"]').setValue("test");
    await wrapper.findComponent('[data-test="save-btn"]').trigger("click");
    await flushPromises();
    expect(storeSpy).toHaveBeenCalledWith({ recovery_email: "test" });
    expect(wrapper.vm.recoveryEmailError).toBe("This recovery email is invalid");
  });

  it("Adds a recovery mail (Fail, Same Email)", async () => {
    await flushPromises();
    const storeSpy = vi.spyOn(usersStore, "patchData");
    mockUsersApi.onPatch("http://localhost:3000/api/users").reply(409);
    await wrapper.findComponent('[data-test="recovery-email-text"]').setValue("test@test.com");
    await wrapper.findComponent('[data-test="save-btn"]').trigger("click");
    await flushPromises();
    expect(storeSpy).toHaveBeenCalledWith({ recovery_email: "test@test.com" });
    expect(wrapper.vm.recoveryEmailError).toBe("This recovery email is already in use");
  });
});
