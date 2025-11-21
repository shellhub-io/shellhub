import { createVuetify } from "vuetify";
import { DOMWrapper, flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { describe, expect, it, vi, beforeEach } from "vitest";
import type { Mock } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import useNamespacesStore from "@admin/store/modules/namespaces";
import NamespaceEdit from "@admin/components/Namespace/NamespaceEdit.vue";
import { IAdminNamespace } from "@admin/interfaces/INamespace";
import { SnackbarInjectionKey } from "@/plugins/snackbar";

const namespace: IAdminNamespace = {
  billing: {
    active: true,
    current_period_end: "",
    customer_id: "",
    payment_failed: null,
    payment_method_id: "",
    price_id: "",
    state: "",
    sub_item_id: "",
    subscription_id: "",
  },
  created_at: "2022-04-13T11:42:49.578Z",
  devices_accepted_count: 1,
  devices_pending_count: 1,
  devices_rejected_count: 0,
  max_devices: 10,
  members: [
    {
      id: "",
      role: "owner",
    },
  ],
  name: "ossystems",
  owner: "ossystems",
  settings: {
    session_record: true,
  },
  tenant_id: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
} as IAdminNamespace;

const mockSnackbar = {
  showSuccess: vi.fn(),
  showError: vi.fn(),
};

describe("Namespace Edit", () => {
  let wrapper: VueWrapper<InstanceType<typeof NamespaceEdit>>;
  let namespacesStore: ReturnType<typeof useNamespacesStore>;

  beforeEach(() => {
    setActivePinia(createPinia());
    namespacesStore = useNamespacesStore();

    namespacesStore.updateNamespace = vi.fn().mockResolvedValue(undefined);
    namespacesStore.fetchNamespaceList = vi.fn().mockResolvedValue(undefined);

    mockSnackbar.showSuccess.mockReset();
    mockSnackbar.showError.mockReset();

    wrapper = mount(NamespaceEdit, {
      global: {
        plugins: [createVuetify()],
        provide: { [SnackbarInjectionKey]: mockSnackbar },
      },
      props: {
        namespace,
        modelValue: true,
      },
    });
  });

  it("Renders the component and dialog", () => {
    expect(wrapper.html()).toMatchSnapshot();

    const dialog = new DOMWrapper(document.body);
    expect(dialog.html()).toMatchSnapshot();
  });

  it("Has the correct initial data", () => {
    expect(wrapper.vm.name).toBe(namespace.name);
    expect(wrapper.vm.maxDevices).toBe(namespace.max_devices);
    expect(wrapper.vm.sessionRecord).toBe(namespace.settings.session_record);

    const dialog = new DOMWrapper(document.body);

    const nameInput = dialog.get<HTMLInputElement>('[data-test="name-text"] input');
    expect(nameInput.element.value).toBe(namespace.name);

    const maxDevicesInput = dialog.get<HTMLInputElement>('[data-test="maxDevices-text"] input');
    expect(maxDevicesInput.element.value).toBe(String(namespace.max_devices));
  });

  it("Calls namespace store and snackbar on form submission with updated values", async () => {
    wrapper.vm.name = "updated-namespace";
    wrapper.vm.maxDevices = 42;
    wrapper.vm.sessionRecord = false;

    await wrapper.vm.submitForm();
    await flushPromises();

    expect(namespacesStore.updateNamespace).toHaveBeenCalledTimes(1);

    const updateNamespaceMock = namespacesStore.updateNamespace as Mock;
    const payload = updateNamespaceMock.mock.calls[0][0];

    expect(payload).toMatchObject({
      name: "updated-namespace",
      max_devices: 42,
      settings: {
        ...namespace.settings,
        session_record: false,
      },
    });

    expect(mockSnackbar.showSuccess).toHaveBeenCalledWith("Namespace updated successfully.");
    expect(mockSnackbar.showError).not.toHaveBeenCalled();
  });

  it("Shows error snackbar when updateNamespace fails", async () => {
    const updateNamespaceMock = namespacesStore.updateNamespace as Mock;
    updateNamespaceMock.mockRejectedValueOnce(new Error("update failed"));

    await wrapper.vm.submitForm();
    await flushPromises();

    expect(namespacesStore.updateNamespace).toHaveBeenCalledTimes(1);
    expect(namespacesStore.fetchNamespaceList).not.toHaveBeenCalled();
    expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to update namespace.");
    expect(mockSnackbar.showSuccess).not.toHaveBeenCalled();
  });
});
