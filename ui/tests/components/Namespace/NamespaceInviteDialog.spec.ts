import { createVuetify } from "vuetify";
import { DOMWrapper, flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import { nextTick } from "vue";
import NamespaceInviteDialog from "@/components/Namespace/NamespaceInviteDialog.vue";
import { namespacesApi, usersApi } from "@/api/http";
import { store, key } from "@/store";
import { router } from "@/router";
import { envVariables } from "@/envVariables";
import { SnackbarPlugin } from "@/plugins/snackbar";

const node = document.createElement("div");
node.setAttribute("id", "app");
document.body.appendChild(node);

type NamespaceInviteDialogWrapper = VueWrapper<InstanceType<typeof NamespaceInviteDialog>>;

describe("Namespace Invite Dialog", () => {
  let wrapper: NamespaceInviteDialogWrapper;

  const vuetify = createVuetify();

  let mockNamespace: MockAdapter;

  let mockUser: MockAdapter;

  const members = [
    {
      id: "507f1f77bcf86cd799439011",
      username: "test",
      role: "owner",
    },
  ];

  const namespaceData = {
    name: "test",
    owner: "test",
    tenant_id: "fake-tenant",
    members,
    settings: {
      session_record: true,
      connection_announcement: "",
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
    tenant: "fake-tenant",
    email: "test@test.com",
    id: "507f1f77bcf86cd799439011",
    role: "owner",
    mfa: {
      enable: false,
      validate: false,
    },
  };

  const session = true;

  beforeEach(async () => {
    await router.push("/accept-invite?sig=test-sig&tenantid=testtentantid");
    store.commit("namespaces/setShowNamespaceInvite", true);
    const el = document.createElement("div");
    document.body.appendChild(el);
    vi.useFakeTimers();
    localStorage.setItem("tenant", "fake-tenant");
    envVariables.isCloud = true;

    mockNamespace = new MockAdapter(namespacesApi.getAxios());
    mockUser = new MockAdapter(usersApi.getAxios());

    mockNamespace.onGet("http://localhost:3000/api/namespaces/fake-tenant-data").reply(200, namespaceData);
    mockUser.onGet("http://localhost:3000/api/users/security").reply(200, session);
    mockUser.onGet("http://localhost:3000/api/auth/user").reply(200, authData);
    mockUser.onGet("http://localhost:3000/api/auth/user").reply(200, authData);

    store.commit("auth/authSuccess", authData);
    store.commit("auth/changeData", authData);
    store.commit("namespaces/setNamespace", namespaceData);
    store.commit("security/setSecurity", session);

    wrapper = mount(NamespaceInviteDialog, {
      attachTo: el,
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

  it("Renders dialog elements with correct data-test attributes", () => {
    const dialog = new DOMWrapper(document.body);

    expect(dialog.find('[data-test="card-dialog"]').exists()).toBe(true);
    expect(dialog.find('[data-test="dialog-title"]').exists()).toBe(true);
    expect(dialog.find('[data-test="dialog-content"]').exists()).toBe(true);
    expect(dialog.find('[data-test="dialog-row"]').exists()).toBe(true);
    expect(dialog.find('[data-test="dialog-col"]').exists()).toBe(true);
    expect(dialog.find('[data-test="dialog-message"]').exists()).toBe(true);
    expect(dialog.find('[data-test="dialog-actions"]').exists()).toBe(true);
    expect(dialog.find('[data-test="decline-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="dialog-spacer"]').exists()).toBe(true);
    expect(dialog.find('[data-test="accept-btn"]').exists()).toBe(true);
  });

  it("Displays the correct modal title and message", () => {
    const dialog = new DOMWrapper(document.body);

    expect(dialog.find('[data-test="dialog-title"]').text()).toBe(wrapper.vm.modalTitle);
    expect(dialog.find('[data-test="dialog-message"]').text()).toBe(wrapper.vm.modalMessage);
  });

  it("Calls declineInvite method when Decline Invitation button is clicked", async () => {
    const declineSpy = vi.spyOn(wrapper.vm, "declineInvite");
    await wrapper.findComponent('[data-test="decline-btn"]').trigger("click");

    await flushPromises();
    expect(declineSpy).toHaveBeenCalled();
  });

  it("Calls acceptInvite method when Accept Invitation button is clicked", async () => {
    const acceptSpy = vi.spyOn(wrapper.vm, "acceptInvite");

    await wrapper.findComponent('[data-test="accept-btn"]').trigger("click");

    await flushPromises();
    expect(acceptSpy).toHaveBeenCalled();
  });

  it("Calls close method when Close button is clicked", async () => {
    wrapper.vm.modalError = true;
    await nextTick();

    const closeSpy = vi.spyOn(wrapper.vm, "close");
    await wrapper.findComponent('[data-test="close-btn"]').trigger("click");

    expect(closeSpy).toHaveBeenCalled();
  });

  it("Handles error state correctly", async () => {
    wrapper.vm.modalError = true;
    await nextTick();
    const dialog = new DOMWrapper(document.body);
    expect(dialog.find('[data-test="error-dialog-actions"]').exists()).toBe(true);
    expect(dialog.find('[data-test="close-btn"]').exists()).toBe(true);
  });

  it("Displays appropriate buttons based on error state", async () => {
    const dialog = new DOMWrapper(document.body);

    // Initially, it should show the Accept and Decline buttons
    expect(dialog.find('[data-test="accept-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="decline-btn"]').exists()).toBe(true);

    // When modalError is true, it should only show the Close button
    wrapper.vm.modalError = true;
    await nextTick();
    expect(dialog.find('[data-test="close-btn"]').exists()).toBe(true);
  });
});
