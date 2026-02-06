import { describe, expect, it, afterEach, vi, beforeEach } from "vitest";
import { VueWrapper, DOMWrapper, flushPromises } from "@vue/test-utils";
import { mountComponent } from "@tests/utils/mount";
import { createAxiosError } from "@tests/utils/axiosError";
import DeviceRename from "@/components/Devices/DeviceRename.vue";
import useDevicesStore from "@/store/modules/devices";

describe("DeviceRename", () => {
  let wrapper: VueWrapper<InstanceType<typeof DeviceRename>>;
  let dialog: DOMWrapper<Element>;
  let devicesStore: ReturnType<typeof useDevicesStore>;

  beforeEach(() => {
    wrapper = mountComponent(DeviceRename, {
      props: { uid: "test-device-uid", name: "test-device-name" },
    });

    devicesStore = useDevicesStore();
    dialog = new DOMWrapper(document.body);
  });

  afterEach(() => {
    wrapper?.unmount();
    document.body.innerHTML = "";
  });

  describe("list item", () => {
    it("renders rename list item", () => {
      expect(wrapper.find('[data-test="rename-icon"]').exists()).toBe(true);
    });

    it("displays Rename text", () => {
      expect(wrapper.find('[data-test="rename-title"]').text()).toBe("Rename");
    });

    it("opens dialog when clicked", async () => {
      await wrapper.find("[data-test='rename-device-button']").trigger("click");
      await flushPromises();

      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      expect(formDialog.props("modelValue")).toBe(true);
    });
  });

  describe("rename dialog", () => {
    it("shows FormDialog with correct props", async () => {
      await wrapper.find("[data-test='rename-device-button']").trigger("click");
      await flushPromises();

      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      expect(formDialog.props("title")).toBe("Rename Device");
      expect(formDialog.props("icon")).toBe("mdi-pencil");
      expect(formDialog.props("confirmText")).toBe("Rename");
      expect(formDialog.props("cancelText")).toBe("Close");
    });

    it("displays text field with device name", async () => {
      await wrapper.find("[data-test='rename-device-button']").trigger("click");
      await flushPromises();

      const textField = dialog.find('[data-test="rename-field"]');
      expect(textField.exists()).toBe(true);
    });

    it("shows initial device name in field", async () => {
      await wrapper.find("[data-test='rename-device-button']").trigger("click");
      await flushPromises();

      const input = dialog.find('[data-test="rename-field"] input').element as HTMLInputElement;
      expect(input.value).toBe("test-device-name");
    });

    it("updates field value when input changes", async () => {
      await wrapper.find("[data-test='rename-device-button']").trigger("click");
      await flushPromises();

      const input = dialog.find('[data-test="rename-field"] input');
      await input.setValue("new-name");
      await flushPromises();

      expect((input.element as HTMLInputElement).value).toBe("new-name");
    });
  });

  describe("device renaming", () => {
    it("calls renameDevice when confirmed", async () => {
      await wrapper.find("[data-test='rename-device-button']").trigger("click");
      await flushPromises();

      const input = dialog.find('[data-test="rename-field"] input');
      await input.setValue("new-device-name");
      await flushPromises();

      const confirmBtn = dialog.find('[data-test="rename-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(devicesStore.renameDevice).toHaveBeenCalledWith({
        uid: "test-device-uid",
        name: { name: "new-device-name" },
      });
    });

    it("emits update event after successful rename", async () => {
      vi.spyOn(devicesStore, "renameDevice").mockResolvedValue();

      await wrapper.find("[data-test='rename-device-button']").trigger("click");
      await flushPromises();

      const input = dialog.find('[data-test="rename-field"] input');
      await input.setValue("new-name");
      await flushPromises();

      const confirmBtn = dialog.find('[data-test="rename-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(wrapper.emitted("update")).toBeTruthy();
    });

    it("closes dialog after successful rename", async () => {
      vi.spyOn(devicesStore, "renameDevice").mockResolvedValue();

      await wrapper.find("[data-test='rename-device-button']").trigger("click");
      await flushPromises();

      const input = dialog.find('[data-test="rename-field"] input');
      await input.setValue("new-name");
      await flushPromises();

      const confirmBtn = dialog.find('[data-test="rename-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(dialog.find(".v-overlay__content").attributes("style")).toContain("display: none");
    });

    it("shows error message when invalid characters are used (400)", async () => {
      vi.spyOn(devicesStore, "renameDevice").mockRejectedValue(createAxiosError(400, "Bad Request"));

      await wrapper.find("[data-test='rename-device-button']").trigger("click");
      await flushPromises();

      const input = dialog.find('[data-test="rename-field"] input');
      await input.setValue("invalid@name");
      await flushPromises();

      const confirmBtn = dialog.find('[data-test="rename-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(dialog.text()).toContain("The characters being used are invalid");
      expect(dialog.find('[data-test="device-rename-dialog"]').exists()).toBe(true);
    });

    it("shows error message when name already exists (409)", async () => {
      vi.spyOn(devicesStore, "renameDevice").mockRejectedValue(createAxiosError(409, "Conflict"));

      await wrapper.find("[data-test='rename-device-button']").trigger("click");
      await flushPromises();

      const input = dialog.find('[data-test="rename-field"] input');
      await input.setValue("existing-name");
      await flushPromises();

      const confirmBtn = dialog.find('[data-test="rename-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(dialog.text()).toContain("The name already exists in the namespace");
      expect(dialog.find('[data-test="device-rename-dialog"]').exists()).toBe(true);
    });

    it("keeps dialog open on error so user can correct input", async () => {
      vi.spyOn(devicesStore, "renameDevice").mockRejectedValue(createAxiosError(400, "Bad Request"));

      await wrapper.find("[data-test='rename-device-button']").trigger("click");
      await flushPromises();

      const input = dialog.find('[data-test="rename-field"] input');
      await input.setValue("invalid@name");
      await flushPromises();

      const confirmBtn = dialog.find('[data-test="rename-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(dialog.find('[data-test="rename-field"]').exists()).toBe(true);
      expect(dialog.find('[data-test="rename-btn"]').exists()).toBe(true);
    });
  });
});
