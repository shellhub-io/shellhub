import { createPinia, setActivePinia } from "pinia";
import { createVuetify } from "vuetify";
import { DOMWrapper, flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import { nextTick } from "vue";
import NamespaceEdit from "@/components/Namespace/NamespaceEdit.vue";
import { namespacesApi } from "@/api/http";
import { SnackbarInjectionKey } from "@/plugins/snackbar";
import useNamespacesStore from "@/store/modules/namespaces";
import { INamespaceMember } from "@/interfaces/INamespace";

type NamespaceEditWrapper = VueWrapper<InstanceType<typeof NamespaceEdit>>;

const mockSnackbar = {
  showSuccess: vi.fn(),
  showError: vi.fn(),
};

describe("Namespace Edit", () => {
  let wrapper: NamespaceEditWrapper;
  setActivePinia(createPinia());
  const namespacesStore = useNamespacesStore();
  const vuetify = createVuetify();

  const mockNamespacesApi = new MockAdapter(namespacesApi.getAxios());

  const members = [
    {
      id: "xxxxxxxx",
      role: "owner" as const,
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
  };

  beforeEach(async () => {
    localStorage.setItem("tenant", "fake-tenant-data");
    mockNamespacesApi.onGet("http://localhost:3000/api/namespaces/fake-tenant-data").reply(200, namespaceData);
    mockNamespacesApi.onGet("http://localhost:3000/api/namespaces?page=1&per_page=10").reply(200, [namespaceData]);

    namespacesStore.currentNamespace = namespaceData;

    wrapper = mount(NamespaceEdit, {
      global: {
        plugins: [vuetify],
        provide: { [SnackbarInjectionKey]: mockSnackbar },
      },
    });
  });

  afterEach(() => {
    wrapper.unmount();
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
    expect(dialog.find('[data-test="title"]').exists()).toBe(true);
    expect(dialog.find('[data-test="close-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="change-connection-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="connection-announcement-text"]').exists()).toBe(true);
  });

  it("Successfully changes connection_announcement data", async () => {
    wrapper.vm.showDialog = true;
    await flushPromises();
    const changeNamespaceData = {
      tenant_id: "fake-tenant-data",
      settings: {
        connection_announcement: "test",
      },
    };

    mockNamespacesApi.onPut("http://localhost:3000/api/namespaces/fake-tenant-data").reply(200, changeNamespaceData);

    await wrapper.findComponent('[data-test="connection-announcement-text"]').setValue("test");

    const storeSpy = vi.spyOn(namespacesStore, "editNamespace");
    await wrapper.findComponent('[data-test="change-connection-btn"]').trigger("click");

    await nextTick();
    await flushPromises();
    expect(storeSpy).toHaveBeenCalledWith(changeNamespaceData);
  });

  it("Fails to change namespace data", async () => {
    wrapper.vm.showDialog = true;
    await flushPromises();
    mockNamespacesApi.onPut("http://localhost:3000/api/namespaces/fake-tenant-data").reply(403);

    await wrapper.findComponent('[data-test="change-connection-btn"]').trigger("click");

    await nextTick();
    await flushPromises();
    expect(mockSnackbar.showError).toHaveBeenCalledWith("An error occurred while updating the connection announcement.");
  });
});
