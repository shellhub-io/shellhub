import { describe, expect, it, afterEach, vi } from "vitest";
import { VueWrapper, flushPromises } from "@vue/test-utils";
import { mountComponent } from "@tests/utils/mount";
import { createCleanRouter } from "@tests/utils/router";
import DeviceLink from "@/components/Devices/DeviceLink.vue";
import { Router } from "vue-router";

describe("DeviceLink", () => {
  let wrapper: VueWrapper<InstanceType<typeof DeviceLink>>;
  let router: Router;

  const mountWrapper = (deviceUid = "test-device-uid", deviceName = "test-device") => {
    router = createCleanRouter();

    wrapper = mountComponent(DeviceLink, {
      global: { plugins: [router] },
      props: { deviceUid, deviceName },
    });
  };

  afterEach(() => { wrapper?.unmount(); });

  describe("button rendering", () => {
    it("renders device link button", () => {
      mountWrapper();

      expect(wrapper.find('[data-test="device-link-button"]').exists()).toBe(true);
    });

    it("displays device name in button", () => {
      mountWrapper("device-123", "my-awesome-device");

      expect(wrapper.text()).toContain("my-awesome-device");
    });

    it("shows device icon", () => {
      mountWrapper();

      const button = wrapper.find('[data-test="device-link-button"]');
      expect(button.html()).toContain("mdi-developer-board");
    });
  });

  describe("navigation", () => {
    it("navigates to device details when clicked", async () => {
      mountWrapper("test-uid-123", "test-device");
      const routerPushSpy = vi.spyOn(router, "push");

      await wrapper.find('[data-test="device-link-button"]').trigger("click");
      await flushPromises();

      expect(routerPushSpy).toHaveBeenCalledWith({ name: "DeviceDetails", params: { identifier: "test-uid-123" } });
    });
  });
});
