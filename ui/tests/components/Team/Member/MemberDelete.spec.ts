import { createVuetify } from "vuetify";
import { DOMWrapper, flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import MemberDelete from "@/components/Team/Member/MemberDelete.vue";
import { namespacesApi, usersApi } from "@/api/http";
import { store, key } from "@/store";
import { router } from "@/router";
import { envVariables } from "@/envVariables";
import { SnackbarInjectionKey } from "@/plugins/snackbar";

type MemberDeleteWrapper = VueWrapper<InstanceType<typeof MemberDelete>>;

const mockSnackbar = {
  showSuccess: vi.fn(),
  showError: vi.fn(),
};

describe("Member Delete", () => {
  let wrapper: MemberDeleteWrapper;

  const vuetify = createVuetify();

  let mockNamespace: MockAdapter;

  let mockUser: MockAdapter;

  const members = [
    {
      id: "xxxxxxxx",
      username: "test",
      role: "owner" as const,
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

  beforeEach(async () => {
    localStorage.setItem("tenant", "fake-tenant-data");
    envVariables.isCloud = true;

    mockNamespace = new MockAdapter(namespacesApi.getAxios());
    mockUser = new MockAdapter(usersApi.getAxios());

    mockNamespace.onGet("http://localhost:3000/api/namespaces/fake-tenant-data").reply(200, namespaceData);
    mockUser.onGet("http://localhost:3000/api/auth/user").reply(200, authData);

    store.commit("auth/authSuccess", authData);
    store.commit("auth/changeData", authData);
    store.commit("namespaces/setNamespace", namespaceData);

    wrapper = mount(MemberDelete, {
      global: {
        plugins: [[store, key], vuetify, router],
        provide: { [SnackbarInjectionKey]: mockSnackbar },
      },
      props: {
        member: members[0],
        hasAuthorization: true,
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
    const dialog = new DOMWrapper(document.body);

    expect(wrapper.findComponent('[data-test="member-delete-dialog-btn"]').exists()).toBe(true);

    await wrapper.findComponent('[data-test="member-delete-dialog-btn"]').trigger("click");

    expect(dialog.find('[data-test="member-delete-card"]').exists()).toBe(true);
    expect(dialog.find('[data-test="member-delete-dialog-title"]').exists()).toBe(true);
    expect(dialog.find('[data-test="member-delete-dialog-text"]').exists()).toBe(true);
    expect(dialog.find('[data-test="member-delete-close-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="member-delete-remove-btn"]').exists()).toBe(true);
  });

  it("Delete Member Error Validation", async () => {
    mockNamespace.onDelete("http://localhost:3000/api/namespaces/fake-tenant-data/members/xxxxxxxx").reply(403);

    const storeSpy = vi.spyOn(store, "dispatch");

    await wrapper.findComponent('[data-test="member-delete-dialog-btn"]').trigger("click");

    await wrapper.findComponent('[data-test="member-delete-remove-btn"]').trigger("click");

    await flushPromises();

    expect(storeSpy).toBeCalledWith("namespaces/removeUser", {
      tenant_id: "fake-tenant-data",
      user_id: "xxxxxxxx",
    });

    expect(mockSnackbar.showError).toBeCalledWith("Failed to remove user from namespace.");
  });

  it("Delete Member Success Validation", async () => {
    mockNamespace.onDelete("http://localhost:3000/api/namespaces/fake-tenant-data/members/xxxxxxxx").reply(200);

    const storeSpy = vi.spyOn(store, "dispatch");

    await wrapper.findComponent('[data-test="member-delete-dialog-btn"]').trigger("click");

    await wrapper.findComponent('[data-test="member-delete-remove-btn"]').trigger("click");

    await flushPromises();

    expect(storeSpy).toBeCalledWith("namespaces/removeUser", {
      tenant_id: "fake-tenant-data",
      user_id: "xxxxxxxx",
    });

    expect(mockSnackbar.showSuccess).toBeCalledWith("Successfully removed user from namespace.");
  });
});
