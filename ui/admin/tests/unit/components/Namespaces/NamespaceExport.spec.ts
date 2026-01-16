import { describe, expect, it, beforeEach, vi, afterEach } from "vitest";
import { DOMWrapper, VueWrapper, flushPromises } from "@vue/test-utils";
import { mountComponent, mockSnackbar } from "@tests/utils/mount";
import { createAxiosError } from "@tests/utils/axiosError";
import { saveAs } from "file-saver";
import useNamespacesStore from "@admin/store/modules/namespaces";
import NamespaceExport from "@admin/components/Namespace/NamespaceExport.vue";

vi.mock("file-saver", () => ({
  saveAs: vi.fn(),
}));

describe("NamespaceExport", () => {
  let wrapper: VueWrapper<InstanceType<typeof NamespaceExport>>;
  let namespacesStore: ReturnType<typeof useNamespacesStore>;

  const mountWrapper = () => {
    wrapper = mountComponent(NamespaceExport, {
      attachTo: document.body,
    });

    namespacesStore = useNamespacesStore();
  };

  const getDialog = () => new DOMWrapper(document.body).find('[role="dialog"]');

  afterEach(() => {
    vi.clearAllMocks();
    wrapper?.unmount();
    document.body.innerHTML = "";
  });

  describe("rendering", () => {
    beforeEach(() => mountWrapper());

    it("renders the export button", () => {
      const exportBtn = wrapper.find('[data-test="namespaces-export-btn"]');
      expect(exportBtn.exists()).toBe(true);
      expect(exportBtn.text()).toContain("Export CSV");
    });

    it("does not show dialog initially", () => {
      expect(getDialog().exists()).toBe(false);
    });
  });

  describe("opening dialog", () => {
    beforeEach(() => mountWrapper());

    it("shows dialog when clicking export button", async () => {
      const exportBtn = wrapper.find('[data-test="namespaces-export-btn"]');
      await exportBtn.trigger("click");
      await flushPromises();

      const dialog = getDialog();
      expect(dialog.exists()).toBe(true);
      expect(dialog.text()).toContain("Export namespaces data");
    });

    it("displays all filter options", async () => {
      const exportBtn = wrapper.find('[data-test="namespaces-export-btn"]');
      await exportBtn.trigger("click");
      await flushPromises();

      const dialog = getDialog();
      expect(dialog.text()).toContain("Namespaces with more than:");
      expect(dialog.text()).toContain("Namespaces with no devices");
      expect(dialog.text()).toContain("Namespace with devices, but no sessions");
    });

    it("has 'more than' filter selected by default", async () => {
      const exportBtn = wrapper.find('[data-test="namespaces-export-btn"]');
      await exportBtn.trigger("click");
      await flushPromises();

      const dialog = getDialog();
      const moreThanRadio = dialog.find('[data-test="radio-more-than"] input');
      expect((moreThanRadio.element as HTMLInputElement).checked).toBe(true);
    });

    it("shows number of devices input enabled by default", async () => {
      const exportBtn = wrapper.find('[data-test="namespaces-export-btn"]');
      await exportBtn.trigger("click");
      await flushPromises();

      const dialog = getDialog();
      const numberInput = dialog.find('[data-test="number-of-devices-input"]');
      expect(numberInput.attributes("disabled")).toBeUndefined();
    });
  });

  describe("filter selection", () => {
    beforeEach(() => mountWrapper());

    it("disables number input when selecting 'no devices' filter", async () => {
      const exportBtn = wrapper.find('[data-test="namespaces-export-btn"]');
      await exportBtn.trigger("click");
      await flushPromises();

      const dialog = getDialog();
      const noDevicesRadio = dialog.find('[data-test="radio-no-devices"] input');
      await noDevicesRadio.setValue(true);
      await flushPromises();

      const numberInput = dialog.find('[data-test="number-of-devices-input"] input');
      expect(numberInput.attributes("disabled")).toBeDefined();
    });

    it("disables number input when selecting 'no sessions' filter", async () => {
      const exportBtn = wrapper.find('[data-test="namespaces-export-btn"]');
      await exportBtn.trigger("click");
      await flushPromises();

      const dialog = getDialog();
      const noSessionsRadio = dialog.find('[data-test="radio-no-sessions"] input');
      await noSessionsRadio.setValue(true);
      await flushPromises();

      const numberInput = dialog.find('[data-test="number-of-devices-input"] input');
      expect(numberInput.attributes("disabled")).toBeDefined();
    });

    it("enables number input when selecting 'more than' filter", async () => {
      const exportBtn = wrapper.find('[data-test="namespaces-export-btn"]');
      await exportBtn.trigger("click");
      await flushPromises();

      const dialog = getDialog();

      // Select another filter first
      const radioNoDevices = dialog.find('[data-test="radio-no-devices"] input');
      await radioNoDevices.setValue(true);
      await flushPromises();

      // Then select "more than" again
      const radioMoreThan = dialog.find('[data-test="radio-more-than"] input');
      await radioMoreThan.setValue(true);
      await flushPromises();

      const numberInput = dialog.find('[data-test="number-of-devices-input"] input');
      expect(numberInput.attributes("disabled")).toBeUndefined();
    });
  });

  describe("form validation", () => {
    beforeEach(() => mountWrapper());

    it("shows error for negative number", async () => {
      const exportBtn = wrapper.find('[data-test="namespaces-export-btn"]');
      await exportBtn.trigger("click");
      await flushPromises();

      const dialog = getDialog();
      const numberInput = dialog.find('[data-test="number-of-devices-input"] input');
      await numberInput.setValue("-1");
      await flushPromises();

      expect(dialog.text()).toContain("this must be greater than or equal to 0");
    });

    it("accepts zero as valid input", async () => {
      const exportBtn = wrapper.find('[data-test="namespaces-export-btn"]');
      await exportBtn.trigger("click");
      await flushPromises();

      const dialog = getDialog();
      const numberInput = dialog.find('[data-test="number-of-devices-input"] input');
      await numberInput.setValue("0");
      await flushPromises();

      expect(dialog.text()).not.toContain("must be greater than or equal to 0");
    });

    it("accepts positive numbers", async () => {
      const exportBtn = wrapper.find('[data-test="namespaces-export-btn"]');
      await exportBtn.trigger("click");
      await flushPromises();

      const dialog = getDialog();
      const numberInput = dialog.find('[data-test="number-of-devices-input"] input');
      await numberInput.setValue("10");
      await flushPromises();

      expect(dialog.text()).not.toContain("must be greater than or equal to 0");
    });

    it("disables export button when validation fails", async () => {
      const exportBtn = wrapper.find('[data-test="namespaces-export-btn"]');
      await exportBtn.trigger("click");
      await flushPromises();

      const dialog = getDialog();
      const numberInput = dialog.find('[data-test="number-of-devices-input"] input');
      await numberInput.setValue("-1");
      await flushPromises();

      const confirmBtn = dialog.find('[data-test="confirm-btn"]');
      expect(confirmBtn.attributes("disabled")).toBeDefined();
    });
  });

  describe("exporting namespaces", () => {
    it("exports with 'more than' filter", async () => {
      mountWrapper();
      const exportBtn = wrapper.find('[data-test="namespaces-export-btn"]');
      await exportBtn.trigger("click");
      await flushPromises();

      const dialog = getDialog();
      const numberInput = dialog.find('[data-test="number-of-devices-input"] input');
      await numberInput.setValue("5");
      await flushPromises();

      const confirmBtn = dialog.find('[data-test="confirm-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(namespacesStore.exportNamespacesToCsv).toHaveBeenCalledWith(expect.any(String));
      expect(saveAs).toHaveBeenCalledWith(
        expect.any(Blob),
        "namespaces_more_than_5_devices.csv",
      );
      expect(mockSnackbar.showSuccess).toHaveBeenCalledWith("Namespaces exported successfully.");
    });

    it("exports with 'no devices' filter", async () => {
      mountWrapper();
      const exportBtn = wrapper.find('[data-test="namespaces-export-btn"]');
      await exportBtn.trigger("click");
      await flushPromises();

      const dialog = getDialog();
      const noDevicesRadio = dialog.find('[data-test="radio-no-devices"] input');
      await noDevicesRadio.setValue(true);
      await flushPromises();

      const confirmBtn = dialog.find('[data-test="confirm-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(namespacesStore.exportNamespacesToCsv).toHaveBeenCalled();
      expect(saveAs).toHaveBeenCalledWith(
        expect.any(Blob),
        "namespaces_no_devices.csv",
      );
      expect(mockSnackbar.showSuccess).toHaveBeenCalledWith("Namespaces exported successfully.");
    });

    it("exports with 'no sessions' filter", async () => {
      mountWrapper();
      const exportBtn = wrapper.find('[data-test="namespaces-export-btn"]');
      await exportBtn.trigger("click");
      await flushPromises();

      const dialog = getDialog();
      const noSessionsRadio = dialog.find('[data-test="radio-no-sessions"] input');
      await noSessionsRadio.setValue(true);
      await flushPromises();

      const confirmBtn = dialog.find('[data-test="confirm-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(namespacesStore.exportNamespacesToCsv).toHaveBeenCalled();
      expect(saveAs).toHaveBeenCalledWith(
        expect.any(Blob),
        "namespaces_with_devices_but_no_sessions.csv",
      );
      expect(mockSnackbar.showSuccess).toHaveBeenCalledWith("Namespaces exported successfully.");
    });

    it("closes dialog after successful export", async () => {
      mountWrapper();
      const exportBtn = wrapper.find('[data-test="namespaces-export-btn"]');
      await exportBtn.trigger("click");
      await flushPromises();

      const dialog = getDialog();
      const confirmBtn = dialog.find('[data-test="confirm-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      const dialogContent = getDialog().find(".v-overlay__content");
      expect(dialogContent.attributes("style")).toContain("display: none;");
    });
  });

  describe("error handling", () => {
    it("shows error snackbar when export fails", async () => {
      mountWrapper();
      vi.mocked(namespacesStore.exportNamespacesToCsv).mockRejectedValueOnce(
        createAxiosError(500, "Internal Server Error"),
      );

      const exportBtn = wrapper.find('[data-test="namespaces-export-btn"]');
      await exportBtn.trigger("click");
      await flushPromises();

      const dialog = getDialog();
      const confirmBtn = dialog.find('[data-test="confirm-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Error exporting namespaces.");
      expect(saveAs).not.toHaveBeenCalled();
    });

    it("keeps dialog open when export fails", async () => {
      mountWrapper();
      vi.mocked(namespacesStore.exportNamespacesToCsv).mockRejectedValueOnce(
        createAxiosError(500, "Internal Server Error"),
      );

      const exportBtn = wrapper.find('[data-test="namespaces-export-btn"]');
      await exportBtn.trigger("click");
      await flushPromises();

      const dialog = getDialog();
      const confirmBtn = dialog.find('[data-test="confirm-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(getDialog().exists()).toBe(true);
    });
  });

  describe("closing dialog", () => {
    beforeEach(() => mountWrapper());

    it("closes dialog when cancel button is clicked", async () => {
      const exportBtn = wrapper.find('[data-test="namespaces-export-btn"]');
      await exportBtn.trigger("click");
      await flushPromises();

      const dialog = getDialog();
      const cancelBtn = dialog.find('[data-test="cancel-btn"]');
      await cancelBtn.trigger("click");
      await flushPromises();

      const dialogContent = getDialog().find(".v-overlay__content");
      expect(dialogContent.attributes("style")).toContain("display: none;");
    });

    it("resets form when dialog is closed and reopened", async () => {
      const exportBtn = wrapper.find('[data-test="namespaces-export-btn"]');
      await exportBtn.trigger("click");
      await flushPromises();

      let dialog = getDialog();
      const numberInput = dialog.find('[data-test="number-of-devices-input"] input');
      await numberInput.setValue("42");
      await flushPromises();

      const cancelBtn = dialog.find('[data-test="cancel-btn"]');
      await cancelBtn.trigger("click");
      await flushPromises();

      // Reopen dialog
      await exportBtn.trigger("click");
      await flushPromises();

      dialog = getDialog();
      const numberInputReopened = dialog.find('[data-test="number-of-devices-input"] input');
      expect((numberInputReopened.element as HTMLInputElement).value).toBe("0");
    });

    it("resets filter selection when dialog is closed and reopened", async () => {
      const exportBtn = wrapper.find('[data-test="namespaces-export-btn"]');
      await exportBtn.trigger("click");
      await flushPromises();

      let dialog = getDialog();
      const noDevicesRadio = dialog.find('[data-test="radio-no-devices"] input');
      await noDevicesRadio.setValue(true);
      await flushPromises();

      const cancelBtn = dialog.find('[data-test="cancel-btn"]');
      await cancelBtn.trigger("click");
      await flushPromises();

      // Reopen dialog
      await exportBtn.trigger("click");
      await flushPromises();

      dialog = getDialog();
      const moreThanRadio = dialog.find('[data-test="radio-more-than"] input');
      expect((moreThanRadio.element as HTMLInputElement).checked).toBe(true);
    });
  });
});
