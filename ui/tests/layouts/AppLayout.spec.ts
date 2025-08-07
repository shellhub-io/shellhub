// AppLayout.spec.ts
import { defineComponent, nextTick } from "vue";
import { mount, flushPromises } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import { beforeEach, afterEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import * as components from "vuetify/components";
import * as directives from "vuetify/directives";
import AppLayout from "@/layouts/AppLayout.vue";
import { store, key } from "@/store";
import { router } from "@/router";
import { SnackbarPlugin } from "@/plugins/snackbar";
import { devicesApi, containersApi } from "@/api/http";
import { envVariables } from "@/envVariables";

let mockDevices: MockAdapter;
let mockContainers: MockAdapter;

describe("App Layout Component", () => {
  let wrapper;

  const vuetify = createVuetify({
    components,
    directives,
  });

  const AppWrapperComponent = defineComponent({
    components: { AppLayout },
    template: `
      <v-app>
        <AppLayout />
      </v-app>
    `,
  });

  beforeEach(() => {
    vi.useFakeTimers();

    envVariables.hasWebEndpoints = true;
    store.dispatch("spinner/setStatus", true);

    mockDevices = new MockAdapter(devicesApi.getAxios());
    mockContainers = new MockAdapter(containersApi.getAxios());

    mockDevices
      .onGet("http://localhost/api/devices?filter=&page=1&per_page=10&status=pending")
      .reply(200);
    mockContainers
      .onGet("http://localhost/api/containers?filter=&page=1&per_page=10&status=pending")
      .reply(200);

    wrapper = mount(AppWrapperComponent, {
      global: {
        plugins: [[store, key], vuetify, router, SnackbarPlugin],
        stubs: {
          "router-link": {
            template: "<a><slot /></a>",
          },
          "router-view": true,
        },

      },
      attachTo: document.body,
    });
  });

  afterEach(() => {
    vi.useRealTimers();
    vi.restoreAllMocks();
    if (wrapper) wrapper.unmount();
  });

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
    await store.dispatch("spinner/setStatus", true);
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

  it("Navigates correctly on item click", async () => {
    const layoutWrapper = wrapper.findComponent(AppLayout);
    const item = layoutWrapper.vm.items[0];
    await layoutWrapper.find(`[data-test="${item.icon}-listItem"]`).trigger("click");
    expect(router.currentRoute.value.path).toBe(item.path);
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
