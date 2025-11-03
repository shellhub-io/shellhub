import { createPinia, setActivePinia } from "pinia";
import { defineComponent, nextTick } from "vue";
import { mount, flushPromises } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import { describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import * as components from "vuetify/components";
import * as directives from "vuetify/directives";
import AppLayout from "@/layouts/AppLayout.vue";
import { router } from "@/router";
import { SnackbarPlugin } from "@/plugins/snackbar";
import { devicesApi, containersApi, namespacesApi } from "@/api/http";
import { envVariables } from "@/envVariables";
import useSpinnerStore from "@/store/modules/spinner";

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

describe("App Layout Component", async () => {
  setActivePinia(createPinia());
  const spinnerStore = useSpinnerStore();
  const vuetify = createVuetify({ components, directives });
  const mockDevicesApi = new MockAdapter(devicesApi.getAxios());
  const mockContainersApi = new MockAdapter(containersApi.getAxios());
  const mockNamespacesApi = new MockAdapter(namespacesApi.getAxios());

  vi.stubGlobal("fetch", vi.fn(async () => Promise.resolve({
    json: () => (cards),
  })));

  const AppWrapperComponent = defineComponent({
    components: { AppLayout },
    template: `
      <v-app>
        <AppLayout />
      </v-app>
    `,
  });

  localStorage.setItem("theme", "dark");
  envVariables.hasWebEndpoints = true;
  envVariables.isCloud = true;
  spinnerStore.status = true;

  mockDevicesApi
    .onGet("http://localhost:3000/api/devices?page=1&per_page=10&status=pending")
    .reply(200);
  mockContainersApi
    .onGet("http://localhost:3000/api/containers?page=1&per_page=10&status=pending")
    .reply(200);
  mockNamespacesApi
    .onGet("http://localhost:3000/api/namespaces?page=1&per_page=30")
    .reply(200, []);

  await router.push("/");
  await router.isReady();

  const wrapper = mount(AppWrapperComponent, {
    global: {
      plugins: [vuetify, router, SnackbarPlugin],
      stubs: {
        "router-link": {
          template: "<a><slot /></a>",
        },
        "router-view": true,
      },

    },
    attachTo: document.body,
  });

  await flushPromises();

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders internal components", () => {
    const layoutWrapper = wrapper.findComponent(AppLayout);
    expect(layoutWrapper.find('[data-test="navigation-drawer"]').exists()).toBe(true);
    expect(layoutWrapper.find('[data-test="drawer-toolbar"]').exists()).toBe(true);
    expect(layoutWrapper.find('[data-test="logo"]').exists()).toBe(true);
    expect(layoutWrapper.find('[data-test="list"]').exists()).toBe(true);
    expect(layoutWrapper.find('[data-test="list-item"]').exists()).toBe(true);
    expect(layoutWrapper.find('[data-test="icon"]').exists()).toBe(true);
    expect(layoutWrapper.find('[data-test="app-bar"]').exists()).toBe(true);
    expect(layoutWrapper.find('[data-test="main"]').exists()).toBe(true);
    expect(layoutWrapper.find('[data-test="container"]').exists()).toBe(true);
    expect(layoutWrapper.find('[data-test="overlay"]').exists()).toBe(true);
    expect(layoutWrapper.find('[data-test="progress-circular"]').exists()).toBe(true);
    expect(layoutWrapper.find('[data-test="userWarning-component"]').exists()).toBe(false);
  });

  it("Renders loading screen", async () => {
    spinnerStore.status = true;
    await flushPromises();

    const layoutWrapper = wrapper.findComponent(AppLayout);
    expect(layoutWrapper.find('[data-test="progress-circular"]').exists()).toBeTruthy();
  });

  it("Renders navigation drawer correctly", async () => {
    const layoutWrapper = wrapper.findComponent(AppLayout);
    layoutWrapper.vm.lgAndUp = !layoutWrapper.vm.lgAndUp;
    await nextTick();
    expect(layoutWrapper.find('[data-test="navigation-drawer"]').isVisible()).toBe(layoutWrapper.vm.lgAndUp);
  });

  it("Renders navigation items from router", () => {
    const layoutWrapper = wrapper.findComponent(AppLayout);
    const { items } = layoutWrapper.vm;

    expect(items.length).toBeGreaterThan(0);
    expect(items[0]).toHaveProperty("icon");
    expect(items[0]).toHaveProperty("title");
    expect(items[0]).toHaveProperty("path");
  });

  it("renders BETA chip for Web Endpoints item", async () => {
    const layoutWrapper = wrapper.findComponent(AppLayout);
    await flushPromises();

    const webEndpointsItem = layoutWrapper.find('[data-test="mdi-web-listItem"]');

    expect(webEndpointsItem.exists()).toBe(true);

    const betaChip = layoutWrapper.find('[data-test="isBeta-chip"]');

    expect(betaChip.exists()).toBe(true);
    expect(betaChip.text().toLowerCase()).toBe("beta");
  });
});
