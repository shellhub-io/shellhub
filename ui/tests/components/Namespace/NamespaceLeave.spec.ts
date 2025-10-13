import { createPinia, setActivePinia } from "pinia";
import { createVuetify } from "vuetify";
import { DOMWrapper, flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import NamespaceLeave from "@/components/Namespace/NamespaceLeave.vue";
import { namespacesApi } from "@/api/http";
import { router } from "@/router";
import { SnackbarInjectionKey } from "@/plugins/snackbar";
import useAuthStore from "@/store/modules/auth";
import useNamespacesStore from "@/store/modules/namespaces";
import { INamespaceMember } from "@/interfaces/INamespace";

const mockSnackbar = {
  showSuccess: vi.fn(),
  showError: vi.fn(),
};

type NamespaceLeaveWrapper = VueWrapper<InstanceType<typeof NamespaceLeave>>;

describe("Namespace Leave", () => {
  let wrapper: NamespaceLeaveWrapper;
  setActivePinia(createPinia());
  const authStore = useAuthStore();
  const namespacesStore = useNamespacesStore();
  const vuetify = createVuetify();

  let mockNamespace: MockAdapter;

  const members = [
    {
      id: "507f1f77bcf86cd799439011",
      role: "administrator" as const,
    },
  ] as INamespaceMember[];

  const namespaceData = {
    billing: null,
    name: "test",
    owner: "test",
    tenant_id: "fake-tenant-data",
    members,
    settings: {
      session_record: true,
      connection_announcement: "",
    },
    max_devices: 3,
    devices_accepted_count: 3,
    devices_rejected_count: 0,
    devices_pending_count: 0,
    created_at: "",
    type: "team" as const,
  };

  beforeEach(async () => {
    localStorage.setItem("tenant", "fake-tenant");
    mockNamespace = new MockAdapter(namespacesApi.getAxios());
    mockNamespace.onGet("http://localhost:3000/api/namespaces/fake-tenant-data").reply(200, namespaceData);

    authStore.role = "administrator";
    namespacesStore.currentNamespace = namespaceData;

    wrapper = mount(NamespaceLeave, {
      global: {
        plugins: [vuetify, router],
        provide: { [SnackbarInjectionKey]: mockSnackbar },
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", async () => {
    wrapper.vm.showDialog = true;
    await flushPromises();
    const dialog = new DOMWrapper(document.body);
    expect(dialog.html()).toMatchSnapshot();
  });

  it("Successfully leaves namespace", async () => {
    wrapper.vm.showDialog = true;
    await flushPromises();

    mockNamespace.onDelete("http://localhost:3000/api/namespaces/fake-tenant/members").reply(200, { token: "fake-token" });

    const storeSpy = vi.spyOn(namespacesStore, "leaveNamespace");
    const routerSpy = vi.spyOn(router, "go").mockImplementation(vi.fn());

    await wrapper.findComponent('[data-test="leave-btn"]').trigger("click");

    await flushPromises();

    expect(storeSpy).toHaveBeenCalledWith("fake-tenant");
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
