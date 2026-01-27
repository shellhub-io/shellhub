import { describe, expect, it, afterEach } from "vitest";
import { VueWrapper } from "@vue/test-utils";
import { mountComponent } from "@tests/utils/mount";
import DeviceIcon from "@/components/Devices/DeviceIcon.vue";

describe("DeviceIcon", () => {
  let wrapper: VueWrapper<InstanceType<typeof DeviceIcon>>;

  const mountWrapper = (icon = "") => {
    wrapper = mountComponent(DeviceIcon, { props: { icon } });
  };

  afterEach(() => { wrapper?.unmount(); });

  describe("default icon", () => {
    it("renders default tux icon when no icon specified", () => {
      mountWrapper();

      expect(wrapper.find('[data-test="device-icon"]').classes()).toContain("fl-tux");
    });

    it("renders default icon for unknown operating system", () => {
      mountWrapper("unknown-os");

      expect(wrapper.find('[data-test="device-icon"]').classes()).toContain("fl-tux");
    });
  });

  describe("specific OS icons", () => {
    it("renders ubuntu icon", () => {
      mountWrapper("ubuntu");

      expect(wrapper.find('[data-test="device-icon"]').classes()).toContain("fl-ubuntu");
    });

    it("renders debian icon", () => {
      mountWrapper("debian");

      expect(wrapper.find('[data-test="device-icon"]').classes()).toContain("fl-debian");
    });

    it("renders fedora icon", () => {
      mountWrapper("fedora");

      expect(wrapper.find('[data-test="device-icon"]').classes()).toContain("fl-fedora");
    });

    it("renders arch icon", () => {
      mountWrapper("arch");

      expect(wrapper.find('[data-test="device-icon"]').classes()).toContain("fl-archlinux");
    });

    it("renders docker icon", () => {
      mountWrapper("docker");

      expect(wrapper.find('[data-test="device-icon"]').classes()).toContain("fl-docker");
    });

    it("renders raspberry-pi icon for raspbian", () => {
      mountWrapper("raspbian");

      expect(wrapper.find('[data-test="device-icon"]').classes()).toContain("fl-raspberry-pi");
    });
  });
});
