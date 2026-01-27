import { describe, expect, it, afterEach, beforeEach, vi } from "vitest";
import { VueWrapper, DOMWrapper, flushPromises } from "@vue/test-utils";
import { mountComponent } from "@tests/utils/mount";
import { createAxiosError } from "@tests/utils/axiosError";
import DeviceDelete from "@/components/Devices/DeviceDelete.vue";
import useDevicesStore from "@/store/modules/devices";
import { Router } from "vue-router";
import { createCleanRouter } from "@tests/utils/router";

describe("DeviceDelete", () => {
  let wrapper: VueWrapper<InstanceType<typeof DeviceDelete>>;
  let dialog: DOMWrapper<Element>;
  let router: Router;
  let devicesStore: ReturnType<typeof useDevicesStore>;

  const mountWrapper = (hasAuthorization = true) => {
    router = createCleanRouter();

    wrapper = mountComponent(DeviceDelete, {
      global: { plugins: [router] },
      props: {
        uid: "test-device-uid",
        variant: "device",
        hasAuthorization,
      },
    });

    devicesStore = useDevicesStore();
    dialog = new DOMWrapper(document.body);
  };

  beforeEach(() => mountWrapper());

  afterEach(() => {
    wrapper?.unmount();
    document.body.innerHTML = "";
  });

  describe("list item", () => {
    it("renders delete list item", () => {
      expect(wrapper.find('[data-test="device-delete-item"]').exists()).toBe(true);
    });

    it("displays remove icon", () => {
      expect(wrapper.find('[data-test="remove-icon"]').exists()).toBe(true);
    });

    it("displays Remove text", () => {
      expect(wrapper.find('[data-test="remove-title"]').text()).toBe("Remove");
    });

    it("opens dialog when clicked", async () => {
      await wrapper.find('[data-test="device-delete-item"]').trigger("click");
      await flushPromises();

      expect(dialog.find('[data-test="delete-device-dialog"]').exists()).toBe(true);
    });

    it("is disabled when hasAuthorization is false", () => {
      wrapper.unmount();
      mountWrapper(false);

      const listItem = wrapper.find('[data-test="device-delete-item"]');
      expect(listItem.classes()).toContain("v-list-item--disabled");
    });
  });

  describe("delete dialog", () => {
    it("shows confirmation dialog with correct title", async () => {
      await wrapper.find('[data-test="device-delete-item"]').trigger("click");
      await flushPromises();

      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(messageDialog.props("title")).toBe("Are you sure?");
    });

    it("shows description with variant", async () => {
      await wrapper.find('[data-test="device-delete-item"]').trigger("click");
      await flushPromises();

      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(messageDialog.props("description")).toContain("device");
      expect(messageDialog.props("description")).toContain("cannot be undone");
    });

    it("displays error icon", async () => {
      await wrapper.find('[data-test="device-delete-item"]').trigger("click");
      await flushPromises();

      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(messageDialog.props("icon")).toBe("mdi-alert");
      expect(messageDialog.props("iconColor")).toBe("error");
    });

    it("shows Remove and Close buttons", async () => {
      await wrapper.find('[data-test="device-delete-item"]').trigger("click");
      await flushPromises();

      expect(dialog.find('[data-test="confirm-btn"]').exists()).toBe(true);
      expect(dialog.find('[data-test="close-btn"]').exists()).toBe(true);
    });

    it("closes dialog when cancel is clicked", async () => {
      await wrapper.find('[data-test="device-delete-item"]').trigger("click");
      await flushPromises();

      await dialog.find('[data-test="close-btn"]').trigger("click");
      await flushPromises();

      expect(dialog.find(".v-overlay__content").attributes("style")).toContain("display: none;");
    });
  });

  describe("device removal", () => {
    it("calls removeDevice when confirmed", async () => {
      await wrapper.find('[data-test="device-delete-item"]').trigger("click");
      await flushPromises();

      await dialog.find('[data-test="confirm-btn"]').trigger("click");
      await flushPromises();

      expect(devicesStore.removeDevice).toHaveBeenCalledWith("test-device-uid");
    });

    it("emits update event after successful removal", async () => {
      await wrapper.find('[data-test="device-delete-item"]').trigger("click");
      await flushPromises();

      await dialog.find('[data-test="confirm-btn"]').trigger("click");
      await flushPromises();

      expect(wrapper.emitted("update")).toBeTruthy();
    });

    it("closes dialog after successful removal", async () => {
      await wrapper.find('[data-test="device-delete-item"]').trigger("click");
      await flushPromises();

      await dialog.find('[data-test="confirm-btn"]').trigger("click");
      await flushPromises();

      expect(dialog.find(".v-overlay__content").attributes("style")).toContain("display: none;");
    });

    it("handles removal error gracefully", async () => {
      vi.mocked(devicesStore.removeDevice).mockRejectedValue(createAxiosError(500, "Internal Server Error"));

      await wrapper.find('[data-test="device-delete-item"]').trigger("click");
      await flushPromises();

      await dialog.find('[data-test="confirm-btn"]').trigger("click");
      await flushPromises();

      expect(dialog.find(".v-overlay__content").attributes("style")).toContain("display: none;");
    });
  });
});
