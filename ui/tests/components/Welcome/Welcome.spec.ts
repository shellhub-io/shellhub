import { createPinia, setActivePinia } from "pinia";
import { DOMWrapper, flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import { expect, describe, it, vi, afterEach } from "vitest";
import Welcome from "@/components/Welcome/Welcome.vue";
import { SnackbarPlugin } from "@/plugins/snackbar";
import useStatsStore from "@/store/modules/stats";

describe("Welcome", () => {
  let wrapper: VueWrapper<InstanceType<typeof Welcome>>;
  const vuetify = createVuetify();
  localStorage.setItem("tenant", "test-tenant");
  setActivePinia(createPinia());
  const statsStore = useStatsStore();
  statsStore.fetchStats = vi.fn().mockResolvedValue({
    registered_devices: 0,
    pending_devices: 0,
    rejected_devices: 0,
    online_devices: 0,
    active_sessions: 0,
  });

  const mountWrapper = () => {
    wrapper = mount(Welcome, { global: { plugins: [vuetify, SnackbarPlugin] } });
  };

  afterEach(() => { wrapper?.unmount(); });

  it("Enables 'Next' (confirm) button when the user sets up a device on step 2", async () => {
    mountWrapper();
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
    localStorage.setItem("namespacesWelcome", "{\"test-tenant\":true}");
    mountWrapper();
    await flushPromises();

    const dialog = new DOMWrapper(document.body);
    expect(dialog.find('[data-test="welcome-window"]').exists()).toBe(false);
  });

  it("Does not render when namespace has devices", async () => {
    statsStore.stats.registered_devices = 1;

    mountWrapper();
    await flushPromises();

    const dialog = new DOMWrapper(document.body);
    expect(dialog.find('[data-test="welcome-window"]').exists()).toBe(false);
  });

  it("Does not render when tenant ID doesn't exist", async () => {
    localStorage.removeItem("tenant");
    mountWrapper();
    await flushPromises();
    const dialog = new DOMWrapper(document.body);
    expect(dialog.find('[data-test="welcome-window"]').exists()).toBe(false);
  });
});
