import { setActivePinia, createPinia } from "pinia";
import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import { nextTick } from "vue";
import SettingProfile from "@/components/Setting/SettingProfile.vue";
import { mfaApi, usersApi } from "@/api/http";
import { router } from "@/router";
import { envVariables } from "@/envVariables";
import { SnackbarPlugin } from "@/plugins/snackbar";
import useAuthStore from "@/store/modules/auth";
import useUsersStore from "@/store/modules/users";

type SettingProfileWrapper = VueWrapper<InstanceType<typeof SettingProfile>>;

const authData = {
  status: "success",
  token: "",
  username: "test",
  name: "test",
  tenant: "fake-tenant-data",
  email: "test@test.com",
  recovery_email: "test2@test.com",
  id: "xxxxxxxx",
  auth_methods: ["local", "saml"],
  role: "owner",
  mfa: {
    enable: false,
    validate: false,
  },
};

const mfaGenerateData = {
  secret: "secret-mfa",
  link: "link-mfa",
  recovery_codes: [
    "HW2wlxV40B",
    "2xsmMUHHHb",
    "DTQgVsaVac",
    "KXPBoXvuWD",
    "QQYTPfotBi",
    "XWiKBEPyb4",
  ],
};

describe("Settings Namespace", () => {
  let wrapper: SettingProfileWrapper;
  setActivePinia(createPinia());
  const authStore = useAuthStore();
  const usersStore = useUsersStore();
  const vuetify = createVuetify();
  const mockUsersApi = new MockAdapter(usersApi.getAxios());
  const mockMfaApi = new MockAdapter(mfaApi.getAxios());

  window.matchMedia = vi.fn().mockImplementation((query) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: vi.fn(),
    removeListener: vi.fn(),
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  }));

  beforeEach(() => {
    envVariables.isCommunity = true; // used in the MFA switch test only
    mockUsersApi.onGet("http://localhost:3000/api/auth/user").reply(200, authData);
    mockMfaApi.onGet("http://localhost:3000/api/user/mfa/generate").reply(200, mfaGenerateData);
    authStore.$patch(authData);
    wrapper = mount(SettingProfile, {
      global: {
        plugins: [vuetify, router, SnackbarPlugin],
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders all data-test elements", () => {
    const dataTests = [
      "account-profile-container",
      "account-profile-card",
      "profile-header",
      "user-icon",
      "profile-title",
      "profile-subtitle",
      "edit-profile-button",
      "profile-details-list",
      "name-field",
      "name-input",
      "username-field",
      "username-input",
      "email-field",
      "email-input",
      "recovery-email-field",
      "recovery-email-input",
      "mfa-card",
      "mfa-text",
      "switch-mfa",
      "delete-account",
      "delete-account-btn",
    ];

    dataTests.forEach((dataTest) => {
      const element = wrapper.find(`[data-test="${dataTest}"]`);
      expect(element.exists()).toBe(true);
    });
  });

  it("Disables the MFA switch for Community accounts", () => {
    const switchMfa = wrapper.find('[data-test="switch-mfa"] input');
    expect(switchMfa.attributes().disabled).toBeDefined();
  });

  it("Successfully changes user data", async () => {
    mockUsersApi.onPatch("http://localhost:3000/api/users").reply(200);

    await wrapper.findComponent('[data-test="edit-profile-button"]').trigger("click");
    await wrapper.findComponent('[data-test="name-input"]').setValue("test");
    await wrapper.findComponent('[data-test="username-input"]').setValue("test");
    await wrapper.findComponent('[data-test="email-input"]').setValue("test@test.com");
    await wrapper.findComponent('[data-test="recovery-email-input"]').setValue("test2@test.com");
    await flushPromises();

    const changeDataSpy = vi.spyOn(usersStore, "patchData");
    await wrapper.findComponent('[data-test="save-changes-button"]').trigger("click");

    await nextTick();
    await flushPromises();
    expect(changeDataSpy).toHaveBeenCalledWith({
      email: "test@test.com",
      name: "test",
      recovery_email: "test2@test.com",
      username: "test",
    });
  });

  it("Fails changes user data", async () => {
    const changeUserData = {
      recovery_email: "test2@test.com",
      email: "test@test.com",
      name: "test",
      username: "test",
    };

    mockUsersApi.onPatch("http://localhost:3000/api/users").reply(401);

    const changeDataSpy = vi.spyOn(usersStore, "patchData");
    await wrapper.findComponent('[data-test="edit-profile-button"]').trigger("click");
    await wrapper.findComponent('[data-test="name-input"]').setValue("test");
    await wrapper.findComponent('[data-test="username-input"]').setValue("test");
    await wrapper.findComponent('[data-test="email-input"]').setValue("test@test.com");
    await wrapper.findComponent('[data-test="recovery-email-input"]').setValue("test2@test.com");
    await flushPromises();

    await wrapper.findComponent('[data-test="save-changes-button"]').trigger("click");

    await nextTick();
    await flushPromises();
    expect(changeDataSpy).toHaveBeenCalledWith(changeUserData);
  });
});
