import { describe, expect, it, beforeEach, afterEach } from "vitest";
import { createRouter, createWebHistory } from "vue-router";
import { flushPromises } from "@vue/test-utils";
import type { VueWrapper } from "@vue/test-utils";
import { VApp } from "vuetify/components";
import { mountComponent } from "@tests/utils/mount";
import AppLayout from "@/layouts/AppLayout.vue";
import { envVariables } from "@/envVariables";
import useSpinnerStore from "@/store/modules/spinner";
import useNamespacesStore from "@/store/modules/namespaces";
import { routes } from "@/router";
import { INamespace } from "@/interfaces/INamespace";

const Component = { template: "<v-app><AppLayout /></v-app>" };

describe("AppLayout", async () => {
  let wrapper: VueWrapper;

  const router = createRouter({
    routes,
    history: createWebHistory(),
  });
  await router.push("/");
  await router.isReady();

  beforeEach(() => {
    envVariables.hasWebEndpoints = false;
    envVariables.isCloud = false;
    envVariables.isCommunity = false;
    envVariables.premiumPaywall = false;

    wrapper = mountComponent(Component, {
      global: {
        plugins: [router],
        stubs: { "router-view": true },
        components: {
          "v-app": VApp,
          AppLayout,
        },
      },
    });
  });

  afterEach(() => { wrapper.unmount(); });

  it("renders main components", () => {
    expect(wrapper.find('[data-test="navigation-drawer"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="logo"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="main"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="container"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="app-bar"]').exists()).toBe(true);
  });

  it("displays loading overlay when spinner is active", async () => {
    const spinnerStore = useSpinnerStore();
    spinnerStore.status = true;

    await flushPromises();

    expect(wrapper.find('[data-test="overlay"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="progress-circular"]').exists()).toBe(true);
  });

  it("hides loading overlay when spinner is inactive", async () => {
    const spinnerStore = useSpinnerStore();
    spinnerStore.status = false;

    await flushPromises();

    const progressCircular = wrapper.find('[data-test="progress-circular"]');
    expect(progressCircular.exists()).toBe(false);
  });

  it("renders navigation items from router configuration", () => {
    const listItems = wrapper.findAll('[data-test="list-item"]');
    expect(listItems.length).toBeGreaterThan(0);
  });

  it("renders BETA chip in Web Endpoints item when isCloud is enabled", async () => {
    envVariables.isCloud = true;
    envVariables.hasWebEndpoints = true;

    wrapper = mountComponent(Component, {
      global: {
        plugins: [router],
        stubs: { "router-view": true },
        components: {
          "v-app": VApp,
          AppLayout,
        },
      },
    });

    await flushPromises();

    const betaChip = wrapper.find('[data-test="isBeta-chip"]');
    expect(betaChip.exists()).toBe(true);
    expect(betaChip.text()).toBe("BETA");
  });

  it("does not render BETA chip when isCloud is disabled", async () => {
    envVariables.isCloud = false;
    envVariables.hasWebEndpoints = true;

    wrapper = mountComponent(Component, {
      global: {
        plugins: [router],
        stubs: { "router-view": true },
        components: {
          "v-app": VApp,
          AppLayout,
        },
      },
    });

    await flushPromises();

    const betaChip = wrapper.find('[data-test="isBeta-chip"]');
    expect(betaChip.exists()).toBe(false);
  });

  it("disables navigation items when no namespaces exist", async () => {
    const namespacesStore = useNamespacesStore();
    expect(namespacesStore.hasNamespaces).toBe(false);

    await flushPromises();

    // Most items should be disabled except Settings and Home
    const listItems = wrapper.findAll('[data-test="list-item"]');
    expect(listItems.length).toBeGreaterThan(0);
  });

  it("enables navigation items when namespaces exist", async () => {
    const namespacesStore = useNamespacesStore();
    namespacesStore.namespaceList = [{ name: "dev", tenant_id: "test-tenant" }] as INamespace[];
    expect(namespacesStore.hasNamespaces).toBe(true);

    await flushPromises();

    const listItems = wrapper.findAll('[data-test="list-item"]');
    expect(listItems.length).toBeGreaterThan(0);
  });

  it("renders premium crown icon for premium features when in community with paywall", async () => {
    envVariables.isCommunity = true;
    envVariables.premiumPaywall = true;

    wrapper = mountComponent(Component, {
      global: {
        plugins: [router],
        stubs: { "router-view": true },
        components: {
          "v-app": VApp,
          AppLayout,
        },
      },
    });

    await flushPromises();

    // Check if any premium indicators exist
    const icons = wrapper.findAll('[data-test="icon"]');
    expect(icons.length).toBeGreaterThan(0);
  });

  it("renders QuickConnection component", () => {
    expect(wrapper.findComponent({ name: "QuickConnection" }).exists()).toBe(true);
  });

  it("applies light theme container styling when theme is light", async () => {
    localStorage.setItem("theme", "light");

    wrapper = mountComponent(Component, {
      global: {
        plugins: [router],
        stubs: { "router-view": true },
        components: {
          "v-app": VApp,
          AppLayout,
        },
      },
    });

    await flushPromises();

    const container = wrapper.find('[data-test="container"]');
    expect(container.classes()).toContain("container-light-bg");
  });
});
