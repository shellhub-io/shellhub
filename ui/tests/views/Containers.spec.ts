import { VueWrapper } from "@vue/test-utils";
import { describe, expect, it, beforeEach, afterEach } from "vitest";
import { mountComponent } from "@tests/utils/mount";
import { createCleanRouter } from "@tests/utils/router";
import Containers from "@/views/Containers.vue";

describe("Containers View", () => {
  let wrapper: VueWrapper<InstanceType<typeof Containers>>;

  const mountWrapper = (showContainers = true) => {
    wrapper = mountComponent(Containers, {
      global: { plugins: [createCleanRouter()] },
      piniaOptions: { initialState: { containers: { showContainers } } },
    });
  };

  afterEach(() => { wrapper?.unmount(); });

  describe("when page loads", () => {
    beforeEach(() => mountWrapper());

    it("renders the page header", () => {
      expect(wrapper.find('[data-test="device-title"]').text()).toContain("Containers");
      expect(wrapper.text()).toContain("Container Management");
    });
  });

  describe("when containers exist", () => {
    beforeEach(() => mountWrapper());

    it("displays the containers table", () => {
      expect(wrapper.find('[data-test="device-table-component"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="no-items-message-component"]').exists()).toBe(false);
    });
  });

  describe("when no containers exist", () => {
    beforeEach(() => mountWrapper(false));
    it("displays the no items message", () => {
      expect(wrapper.find('[data-test="no-items-message-component"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="device-table-component"]').exists()).toBe(false);
    });

    it("shows information about Docker connectors", () => {
      const noItemsMessage = wrapper.find('[data-test="no-items-message-component"]');
      expect(noItemsMessage.text()).toContain("configure a Docker Connector");
    });
  });
});
