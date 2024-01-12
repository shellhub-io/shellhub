import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import { nextTick } from "vue";
import SettingProfile from "@/components/Setting/SettingProfile.vue";
import { namespacesApi, usersApi } from "@/api/http";
import { store, key } from "@/store";
import { router } from "@/router";
import { envVariables } from "@/envVariables";
import { SnackbarPlugin } from "@/plugins/snackbar";

type SettingProfileWrapper = VueWrapper<InstanceType<typeof SettingProfile>>;

describe("Settings Namespace", () => {
  let wrapper: SettingProfileWrapper;

  const vuetify = createVuetify();

  let mockNamespace: MockAdapter;

  let mockUser: MockAdapter;

  const members = [
    {
      id: "xxxxxxxx",
      username: "test",
      role: "owner",
    },
  ];

  const namespaceData = {
    name: "test",
    owner: "test",
    tenant_id: "fake-tenant-data",
    members,
    settings: {
      session_record: true,
    },
    max_devices: 3,
    devices_count: 3,
    created_at: "",
  };

  const authData = {
    status: "success",
    token: "",
    user: "test",
    name: "test",
    tenant: "fake-tenant-data",
    email: "test@test.com",
    id: "xxxxxxxx",
    role: "owner",
    mfa: {
      enable: false,
      validate: false,
    },
  };

  const session = true;

  beforeEach(async () => {
    vi.useFakeTimers();
    localStorage.setItem("tenant", "fake-tenant-data");
    envVariables.isCloud = true;

    mockNamespace = new MockAdapter(namespacesApi.getAxios());
    mockUser = new MockAdapter(usersApi.getAxios());

    mockNamespace.onGet("http://localhost:3000/api/namespaces/fake-tenant-data").reply(200, namespaceData);
    mockUser.onGet("http://localhost:3000/api/users/security").reply(200, session);
    mockUser.onGet("http://localhost:3000/api/auth/user").reply(200, authData);

    store.commit("auth/authSuccess", authData);
    store.commit("auth/changeData", authData);
    store.commit("namespaces/setNamespace", namespaceData);
    store.commit("security/setSecurity", session);

    wrapper = mount(SettingProfile, {
      global: {
        plugins: [[store, key], vuetify, router, SnackbarPlugin],
        config: {
          errorHandler: () => { /* ignore global error handler */ },
        },
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Data is defined", () => {
    expect(wrapper.vm.$data).toBeDefined();
  });

  it("Renders components", async () => {
    expect(wrapper.find('[data-test="account-header"]').exists());
    expect(wrapper.find('[data-test="change-data-btn"]').exists());
    expect(wrapper.find('[data-test="cancel-btn"]').exists());
    expect(wrapper.find('[data-test="update-user-btn"]').exists());
    expect(wrapper.find('[data-test="name-text"]').exists());
    expect(wrapper.find('[data-test="username-text"]').exists());
    expect(wrapper.find('[data-test="email-text"]').exists());
    expect(wrapper.find('[data-test="password-header"]').exists());
    expect(wrapper.find('[data-test="change-password-btn"]').exists());
    expect(wrapper.find('[data-test="cancel-password-btn"]').exists());
    expect(wrapper.find('[data-test="update-password-btn"]').exists());
    expect(wrapper.find('[data-test="password-text"]').exists());
    expect(wrapper.find('[data-test="newPassword-text"]').exists());
    expect(wrapper.find('[data-test="confirmNewPassword-text"]').exists());
  });

  it("Successfully changes user data", async () => {
    const changeUserData = {
      email: "test@test.com",
      name: "test",
      username: "test",
      id: "xxxxxxxx",
    };

    mockUser.onPatch("http://localhost:3000/api/users/xxxxxxxx/data").reply(200);

    await wrapper.findComponent('[data-test="change-data-btn"]').trigger("click");
    await wrapper.findComponent('[data-test="name-text"]').setValue("test");
    await wrapper.findComponent('[data-test="username-text"]').setValue("test");
    await wrapper.findComponent('[data-test="email-text"]').setValue("test@test.com");

    const changeDataSpy = vi.spyOn(store, "dispatch");
    await wrapper.findComponent('[data-test="update-user-btn"]').trigger("click");

    vi.runOnlyPendingTimers();

    await nextTick();
    await flushPromises();
    expect(changeDataSpy).toHaveBeenCalledWith("users/patchData", changeUserData);
  });

  it("Successfully changes user password", async () => {
    const changePasswordData = {
      currentPassword: "test",
      newPassword: "test123",
      id: "xxxxxxxx",
    };

    await wrapper.findComponent('[data-test="change-password-btn"]').trigger("click");

    mockUser.onPatch("http://localhost:3000/api/users/xxxxxxxx/password").reply(200);

    const changePasswordSpy = vi.spyOn(store, "dispatch");
    await wrapper.findComponent('[data-test="password-text"]').setValue("test");
    await wrapper.findComponent('[data-test="newPassword-text"]').setValue("test123");
    await wrapper.findComponent('[data-test="confirmNewPassword-text"]').setValue("test123");

    await wrapper.findComponent('[data-test="update-password-btn"]').trigger("click");

    vi.runOnlyPendingTimers();

    await nextTick();
    await flushPromises();
    expect(changePasswordSpy).toHaveBeenCalledWith("users/patchPassword", changePasswordData);
  });

  it("Fails changes user data", async () => {
    const changeUserData = {
      email: "test@test.com",
      name: "test",
      username: "test",
      id: "xxxxxxxx",
    };

    mockUser.onPatch("http://localhost:3000/api/users/xxxxxxxx/data").reply(401);

    const changeDataSpy = vi.spyOn(store, "dispatch");
    await wrapper.findComponent('[data-test="change-data-btn"]').trigger("click");
    await wrapper.findComponent('[data-test="name-text"]').setValue("test");
    await wrapper.findComponent('[data-test="username-text"]').setValue("test");
    await wrapper.findComponent('[data-test="email-text"]').setValue("test@test.com");
    await wrapper.findComponent('[data-test="update-user-btn"]').trigger("click");

    vi.runOnlyPendingTimers();

    await nextTick();
    await flushPromises();
    expect(changeDataSpy).toHaveBeenCalledWith("users/patchData", changeUserData);
  });

  it("Fails changes user password", async () => {
    const changePasswordData = {
      currentPassword: "test",
      newPassword: "test123",
      id: "xxxxxxxx",
    };

    await wrapper.findComponent('[data-test="change-password-btn"]').trigger("click");

    mockUser.onPatch("http://localhost:3000/api/users/xxxxxxxx/password").reply(401);

    const changePasswordSpy = vi.spyOn(store, "dispatch");
    await wrapper.findComponent('[data-test="password-text"]').setValue("test");
    await wrapper.findComponent('[data-test="newPassword-text"]').setValue("test123");
    await wrapper.findComponent('[data-test="confirmNewPassword-text"]').setValue("test123");

    await wrapper.findComponent('[data-test="update-password-btn"]').trigger("click");

    vi.runOnlyPendingTimers();

    await nextTick();
    await flushPromises();
    expect(changePasswordSpy).toHaveBeenCalledWith("users/patchPassword", changePasswordData);
  });
});
