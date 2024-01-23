import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, afterEach, describe, expect, it, vi } from "vitest";
import { nextTick } from "vue";
import AppLayout from "@/layouts/AppLayout.vue";
import { store, key } from "@/store";
import { router } from "@/router";
import { SnackbarPlugin } from "@/plugins/snackbar";

type AppLayoutWrapper = VueWrapper<InstanceType<typeof AppLayout>>;

describe("App Layout Component", () => {
  let wrapper: AppLayoutWrapper;
  const vuetify = createVuetify();

  beforeEach(() => {
    vi.useFakeTimers();

    wrapper = mount(AppLayout, {
      global: {
        plugins: [[store, key], vuetify, router, SnackbarPlugin],
      },
    });
  });

  afterEach(() => {
    vi.useRealTimers();
    vi.restoreAllMocks();
    wrapper.unmount();
  });

  it("Is a Vue instance", () => {
    // Test if the wrapper represents a Vue instance
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    // Test if the component renders as expected
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Data is defined", () => {
    // Test if the component's data is defined
    expect(wrapper.vm.$data).toBeDefined();
  });

  it("Renders internal components", () => {
    expect(wrapper.find('[data-test="navigation-drawer"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="app-bar-title"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="logo"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="list"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="list-item"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="icon"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="app-bar"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="main"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="container"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="overlay"]').exists()).toBe(false);
    expect(wrapper.find('[data-test="progress-circular"]').exists()).toBe(false);
    expect(wrapper.find('[data-test="userWarning-component"]').exists()).toBe(false);
  });

  it("Renders loading screen", async () => {
    await store.dispatch("spinner/setStatus", true);

    await flushPromises();

    expect(wrapper.find('[data-test="progress-circular"]').exists()).toBeTruthy();
  });

  it("Renders navigation drawer correctly", async () => {
    // Simulate a state change and check if the navigation drawer visibility changes accordingly
    wrapper.vm.lgAndUp = !wrapper.vm.lgAndUp;
    await nextTick();
    expect(wrapper.find('[data-test="navigation-drawer"]').isVisible()).toBe(wrapper.vm.lgAndUp);
  });

  it("Navigates correctly on item click", async () => {
    // Test if clicking on a navigation item navigates to the correct path
    const item = wrapper.vm.items[0];
    await wrapper.find(`[data-test="${item.icon}-listItem"]`).trigger("click");
    expect(router.currentRoute.value.path).toBe(item.path);
  });
});
