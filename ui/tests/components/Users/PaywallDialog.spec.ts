import { setActivePinia, createPinia } from "pinia";
import { flushPromises, DOMWrapper, mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import { expect, describe, it, beforeEach, vi } from "vitest";
import PaywallDialog from "@/components/User/PaywallDialog.vue";
import { SnackbarPlugin } from "@/plugins/snackbar";

const cards = [
  {
    title: "ShellHub Cloud",
    features: [
      "Protection Against DDoS Attacks",
      "Session record and playback",
      "Managing Firewall Rules",
      "Secure remote communication",
    ],
    button: {
      link: "https://www.shellhub.io/pricing",
      label: "Pricing",
    },
  },
  {
    title: "ShellHub Enterprise",
    features: [
      "Dedicated server for each customer",
      "Supports up to thousands of devices",
      "Reduced maintenance cost",
    ],
    button: {
      link: "https://www.shellhub.io/pricing",
      label: "Get a quote",
    },
  },
];

describe("PaywallDialog", () => {
  let wrapper: VueWrapper<InstanceType<typeof PaywallDialog>>;
  setActivePinia(createPinia());
  const vuetify = createVuetify();

  vi.stubGlobal("fetch", vi.fn(async () => Promise.resolve({
    json: () => (cards),
  })));

  beforeEach(() => {
    wrapper = mount(PaywallDialog, {
      global: {
        plugins: [vuetify, SnackbarPlugin],
      },
    });
    wrapper.vm.showDialog = true;
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders the component table", async () => {
    const dialog = new DOMWrapper(document.body);
    await flushPromises();

    expect(dialog.find('[data-test="paywall-features-dialog"]').exists()).toBe(true);
    expect(dialog.find('[data-test="items-row"]').exists()).toBe(true);
  });

  it("Renders the component table with a successful request to get card infos", async () => {
    const dialog = new DOMWrapper(document.body);

    await flushPromises();
    expect(dialog.find('[data-test="item-card-0"]').exists()).toBe(true);
    expect(dialog.find('[data-test="item-card-1"]').exists()).toBe(true);

    expect(dialog.find('[data-test="item-title-0"]').exists()).toBe(true);
    expect(dialog.find('[data-test="item-title-1"]').exists()).toBe(true);

    expect(dialog.find('[data-test="item-content-0"]').exists()).toBe(true);
    expect(dialog.find('[data-test="item-content-1"]').exists()).toBe(true);

    expect(dialog.find('[data-test="item-content-row-0-0"]').exists()).toBe(true);
    expect(dialog.find('[data-test="item-content-row-0-1"]').exists()).toBe(true);
    expect(dialog.find('[data-test="item-content-row-1-0"]').exists()).toBe(true);
    expect(dialog.find('[data-test="item-content-row-1-1"]').exists()).toBe(true);

    expect(dialog.find('[data-test="pricing-btn-0"]').exists()).toBe(true);
    expect(dialog.find('[data-test="pricing-btn-1"]').exists()).toBe(true);

    expect(dialog.find('[data-test="item-actions-0"]').exists()).toBe(true);
    expect(dialog.find('[data-test="item-actions-1"]').exists()).toBe(true);
    expect(dialog.find('[data-test="no-link-available-btn"]').exists()).toBe(false);
  });

  it("Renders the component table with a successful request to get card infos", () => {
    const dialog = new DOMWrapper(document.body);
    expect(dialog.find('[data-test="no-link-available-btn"]').exists()).toBe(true);
  });
});
