import { createPinia, setActivePinia } from "pinia";
import { DOMWrapper, flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import { expect, describe, it, beforeEach, vi, afterEach } from "vitest";
import Welcome from "@/components/Welcome/Welcome.vue";
import { SnackbarPlugin } from "@/plugins/snackbar";
import useNamespacesStore from "@/store/modules/namespaces";
import useStatsStore from "@/store/modules/stats";

const mockNamespace = {
  tenant_id: "test-tenant",
  name: "test-namespace",
  members: [],
  max_devices: 10,
  owner: "owner-id",
  created_at: "",
  settings: {
    session_record: false,
  },
  devices_accepted_count: 0,
  devices_pending_count: 0,
  devices_rejected_count: 0,
  billing: null,
  type: "personal" as const,
};

describe("Welcome", () => {
  let wrapper: VueWrapper<InstanceType<typeof Welcome>>;
  const vuetify = createVuetify();
  setActivePinia(createPinia());
  const namespacesStore = useNamespacesStore();
  const statsStore = useStatsStore();

  beforeEach(() => {
    vi.spyOn(Storage.prototype, "getItem").mockReturnValue("{}");
    vi.spyOn(Storage.prototype, "setItem");
    namespacesStore.currentNamespace = mockNamespace;
    namespacesStore.namespaceList = [mockNamespace];
    statsStore.stats = {
      registered_devices: 0,
      pending_devices: 0,
      rejected_devices: 0,
      online_devices: 0,
      active_sessions: 0,
    };

    wrapper = mount(Welcome, { global: { plugins: [vuetify, SnackbarPlugin] } });
  });

  afterEach(() => { wrapper.unmount(); });

  it("Enables 'Next' (confirm) button when the user sets up a device on step 2", async () => {
    await flushPromises();

    wrapper.vm.currentStep = 2;
    wrapper.vm.hasDeviceDetected = true;
    wrapper.vm.showDialog = true;

    await flushPromises();

    const dialog = new DOMWrapper(document.body);
    const confirmButton = dialog.find('[data-test="confirm-btn"]');
    expect(confirmButton.exists()).toBe(true);
    expect((confirmButton.element as HTMLButtonElement).disabled).toBe(false);
  });

  it("Does not render when namespace has already been shown", async () => {
    vi.spyOn(Storage.prototype, "getItem").mockReturnValue('{"test-tenant":true}');

    await flushPromises();

    const dialog = new DOMWrapper(document.body);
    expect(dialog.find('[data-test="welcome-window"]').exists()).toBe(false);
  });

  it("Does not render when namespace has devices", async () => {
    statsStore.stats.registered_devices = 1;

    await flushPromises();

    const dialog = new DOMWrapper(document.body);
    expect(dialog.find('[data-test="welcome-window"]').exists()).toBe(false);
  });

  it("Does not render when hasNamespaces is false", async () => {
    namespacesStore.namespaceList = [];
    await flushPromises();
    const dialog = new DOMWrapper(document.body);
    expect(dialog.find('[data-test="welcome-window"]').exists()).toBe(false);
  });
});
