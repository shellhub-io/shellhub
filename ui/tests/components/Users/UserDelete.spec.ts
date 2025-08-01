import { createVuetify } from "vuetify";
import { DOMWrapper, flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import UserDelete from "@/components/User/UserDelete.vue";
import { usersApi } from "@/api/http";
import { store, key } from "@/store";
import { router } from "@/router";
import { envVariables } from "@/envVariables";
import { SnackbarInjectionKey } from "@/plugins/snackbar";

const mockSnackbar = {
  showSuccess: vi.fn(),
  showError: vi.fn(),
};

type UserDeleteWrapper = VueWrapper<InstanceType<typeof UserDelete>>;

describe("User Delete", () => {
  let wrapper: UserDeleteWrapper;

  const vuetify = createVuetify();

  let mockUser: MockAdapter;

  const authData = {
    status: "success",
    token: "",
    user: "test",
    name: "test",
    tenant: "fake-tenant",
    email: "test@test.com",
    id: "507f1f77bcf86cd799439011",
    role: "owner",
    mfa: {
      enable: false,
      validate: false,
    },
  };

  beforeEach(async () => {
    localStorage.setItem("tenant", "fake-tenant");
    envVariables.isCloud = true;

    mockUser = new MockAdapter(usersApi.getAxios());

    mockUser.onGet("http://localhost:3000/api/auth/user").reply(200, authData);

    store.commit("auth/authSuccess", authData);
    store.commit("auth/changeData", authData);

    wrapper = mount(UserDelete, {
      global: {
        plugins: [[store, key], vuetify, router],
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

  it("Renders components", async () => {
    wrapper.vm.showDialog = true;
    const dialog = new DOMWrapper(document.body);
    await flushPromises();
    expect(dialog.find('[data-test="user-delete-dialog"]').exists()).toBe(true);
    expect(dialog.find('[data-test="title"]').exists()).toBe(true);
    expect(dialog.find('[data-test="subtitle"]').exists()).toBe(true);
    expect(dialog.find('[data-test="close-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="delete-user-btn"]').exists()).toBe(true);
  });

  it("Successfully Delete User", async () => {
    mockUser.onDelete("http://localhost:3000/api/user").reply(200);

    const StoreSpy = vi.spyOn(store, "dispatch");

    wrapper.vm.showDialog = true;
    await flushPromises();

    await wrapper.findComponent('[data-test="delete-user-btn"]').trigger("click");

    await flushPromises();
    expect(StoreSpy).toHaveBeenCalledWith("auth/deleteUser");
  });

  it("Fails to add Delete User", async () => {
    mockUser.onDelete("http://localhost:3000/api/user").reply(400);

    wrapper.vm.showDialog = true;
    await flushPromises();

    await wrapper.findComponent('[data-test="delete-user-btn"]').trigger("click");
    await flushPromises();
    expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to delete account.");
  });
});
