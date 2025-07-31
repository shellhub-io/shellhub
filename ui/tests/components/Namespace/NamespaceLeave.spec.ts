import { createPinia, setActivePinia } from "pinia";
import { createVuetify } from "vuetify";
import { DOMWrapper, flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import NamespaceLeave from "@/components/Namespace/NamespaceLeave.vue";
import { namespacesApi } from "@/api/http";
import { store, key } from "@/store";
import { router } from "@/router";
import { SnackbarInjectionKey } from "@/plugins/snackbar";
import useAuthStore from "@/store/modules/auth";

const mockSnackbar = {
  showSuccess: vi.fn(),
  showError: vi.fn(),
};

type NamespaceLeaveWrapper = VueWrapper<InstanceType<typeof NamespaceLeave>>;

describe("Namespace Leave", () => {
  let wrapper: NamespaceLeaveWrapper;
  setActivePinia(createPinia());
  const authStore = useAuthStore();
  const vuetify = createVuetify();

  let mockNamespace: MockAdapter;

  const members = [
    {
      id: "507f1f77bcf86cd799439011",
      username: "test",
      role: "administrator",
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

  beforeEach(async () => {
    localStorage.setItem("tenant", "fake-tenant");
    mockNamespace = new MockAdapter(namespacesApi.getAxios());
    mockNamespace.onGet("http://localhost:3000/api/namespaces/fake-tenant-data").reply(200, namespaceData);
    store.commit("namespaces/setNamespace", namespaceData);
    authStore.role = "administrator";

    wrapper = mount(NamespaceLeave, {
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
    const dialog = new DOMWrapper(document.body);
    wrapper.vm.showDialog = true;
    await flushPromises();

    expect(dialog.find('[data-test="namespace-leave-dialog"]').exists()).toBe(true);
    expect(dialog.find('[data-test="title"]').exists()).toBe(true);
    expect(dialog.find('[data-test="subtitle"]').exists()).toBe(true);
    expect(dialog.find('[data-test="close-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="leave-btn"]').exists()).toBe(true);
  });

  it("Successfully leaves namespace", async () => {
    wrapper.vm.showDialog = true;
    await flushPromises();

    mockNamespace.onDelete("http://localhost:3000/api/namespaces/fake-tenant/members").reply(200, { token: "fake-token" });

    const storeSpy = vi.spyOn(store, "dispatch");
    const routerSpy = vi.spyOn(router, "go").mockImplementation(vi.fn());

    await wrapper.findComponent('[data-test="leave-btn"]').trigger("click");

    await flushPromises();

    expect(storeSpy).toHaveBeenCalledWith("namespaces/leave", "fake-tenant");
    expect(routerSpy).toHaveBeenCalledWith(0);
  });

  it("Fails to Edit Api Key", async () => {
    wrapper.vm.showDialog = true;
    await flushPromises();

    mockNamespace.onDelete("http://localhost:3000/api/namespaces/fake-tenant/members").reply(400);

    await wrapper.findComponent('[data-test="leave-btn"]').trigger("click");

    await flushPromises();

    expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to leave the namespace.");
  });
});
