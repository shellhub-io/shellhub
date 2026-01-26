import { describe, expect, it, afterEach, beforeEach } from "vitest";
import { VueWrapper } from "@vue/test-utils";
import { mountComponent } from "@tests/utils/mount";
import { createCleanRouter } from "@tests/utils/router";
import Container from "@/components/Containers/Container.vue";
import { Router } from "vue-router";
import useContainersStore from "@/store/modules/containers";

describe("Container", () => {
  let wrapper: VueWrapper<InstanceType<typeof Container>>;
  let containersStore: ReturnType<typeof useContainersStore>;
  let router: Router;

  beforeEach(() => {
    router = createCleanRouter();
    wrapper = mountComponent(Container, { global: { plugins: [router] } });
    containersStore = useContainersStore();
  });

  afterEach(() => { wrapper?.unmount(); });

  describe("navigation tabs", () => {
    it("renders three state tabs", () => {
      const tabs = wrapper.findAll('[data-test="container-state-btn"]');
      expect(tabs.length).toBeGreaterThanOrEqual(3);
    });

    it("displays Accepted tab", () => {
      expect(wrapper.text()).toContain("Accepted");
    });

    it("displays Pending tab", () => {
      expect(wrapper.text()).toContain("Pending");
    });

    it("displays Rejected tab", () => {
      expect(wrapper.text()).toContain("Rejected");
    });
  });

  describe("search functionality", () => {
    it("renders search input field", () => {
      const searchField = wrapper.find('[data-test="search-text"]');
      expect(searchField.exists()).toBe(true);
    });

    it("updates container filter when search input changes", async () => {
      const searchField = wrapper.find('[data-test="search-text"] input');
      await searchField.setValue("test-container");

      expect(containersStore.containerListFilter).toBeDefined();
    });

    it("clears filter when search input is empty", async () => {
      const searchField = wrapper.find('[data-test="search-text"] input');
      await searchField.setValue("test");
      await searchField.setValue("");

      expect(containersStore.containerListFilter).toBeUndefined();
    });
  });

  describe("router view", () => {
    it("renders router-view for nested routes", () => {
      expect(wrapper.findComponent({ name: "RouterView" }).exists()).toBe(true);
    });
  });
});
