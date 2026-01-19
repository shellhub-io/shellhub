import { describe, expect, it, beforeEach, vi, afterEach } from "vitest";
import { DOMWrapper, VueWrapper, flushPromises } from "@vue/test-utils";
import { mountComponent, mockSnackbar } from "@tests/utils/mount";
import { createAxiosError } from "@tests/utils/axiosError";
import { saveAs } from "file-saver";
import useUsersStore from "@admin/store/modules/users";
import UserExport from "@admin/components/User/UserExport.vue";

vi.mock("file-saver", () => ({
  saveAs: vi.fn(),
}));

describe("UserExport", () => {
  let wrapper: VueWrapper<InstanceType<typeof UserExport>>;
  let usersStore: ReturnType<typeof useUsersStore>;

  const mountWrapper = () => {
    wrapper = mountComponent(UserExport, {
      attachTo: document.body,
    });

    usersStore = useUsersStore();
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
      const exportBtn = wrapper.find('[data-test="users-export-btn"]');
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
      const exportBtn = wrapper.find('[data-test="users-export-btn"]');
      await exportBtn.trigger("click");
      await flushPromises();

      const dialog = getDialog();
      expect(dialog.exists()).toBe(true);
      expect(dialog.text()).toContain("Export users data");
    });

    it("displays all filter options", async () => {
      const exportBtn = wrapper.find('[data-test="users-export-btn"]');
      await exportBtn.trigger("click");
      await flushPromises();

      const dialog = getDialog();
      expect(dialog.text()).toContain("Users with more than:");
      expect(dialog.text()).toContain("Users with exactly:");
    });

    it("has 'more than' filter selected by default", async () => {
      const exportBtn = wrapper.find('[data-test="users-export-btn"]');
      await exportBtn.trigger("click");
      await flushPromises();

      const dialog = getDialog();
      const moreThanRadio = dialog.find('[data-test="radio-more-than"] input');
      expect((moreThanRadio.element as HTMLInputElement).checked).toBe(true);
    });

    it("shows number of namespaces input with default value 0", async () => {
      const exportBtn = wrapper.find('[data-test="users-export-btn"]');
      await exportBtn.trigger("click");
      await flushPromises();

      const dialog = getDialog();
      const numberInput = dialog.find('[data-test="number-of-namespaces-input"] input');
      expect((numberInput.element as HTMLInputElement).value).toBe("0");
    });
  });

  describe("filter selection", () => {
    beforeEach(() => mountWrapper());

    it("allows selecting 'exactly' filter", async () => {
      const exportBtn = wrapper.find('[data-test="users-export-btn"]');
      await exportBtn.trigger("click");
      await flushPromises();

      const dialog = getDialog();
      const exactlyRadio = dialog.find('[data-test="radio-exactly"] input');
      await exactlyRadio.setValue(true);
      await flushPromises();

      expect((exactlyRadio.element as HTMLInputElement).checked).toBe(true);
    });

    it("allows switching back to 'more than' filter", async () => {
      const exportBtn = wrapper.find('[data-test="users-export-btn"]');
      await exportBtn.trigger("click");
      await flushPromises();

      const dialog = getDialog();

      // Select exactly filter first
      const exactlyRadio = dialog.find('[data-test="radio-exactly"] input');
      await exactlyRadio.setValue(true);
      await flushPromises();

      // Then select "more than" again
      const moreThanRadio = dialog.find('[data-test="radio-more-than"] input');
      await moreThanRadio.setValue(true);
      await flushPromises();

      expect((moreThanRadio.element as HTMLInputElement).checked).toBe(true);
    });
  });

  describe("form validation", () => {
    beforeEach(() => mountWrapper());

    it("shows error for negative number", async () => {
      const exportBtn = wrapper.find('[data-test="users-export-btn"]');
      await exportBtn.trigger("click");
      await flushPromises();

      const dialog = getDialog();
      const numberInput = dialog.find('[data-test="number-of-namespaces-input"] input');
      await numberInput.setValue("-1");
      await flushPromises();

      expect(dialog.text()).toContain("this must be greater than or equal to 0");
    });

    it("accepts zero as valid input", async () => {
      const exportBtn = wrapper.find('[data-test="users-export-btn"]');
      await exportBtn.trigger("click");
      await flushPromises();

      const dialog = getDialog();
      const numberInput = dialog.find('[data-test="number-of-namespaces-input"] input');
      await numberInput.setValue("0");
      await flushPromises();

      expect(dialog.text()).not.toContain("must be greater than or equal to 0");
    });

    it("accepts positive numbers", async () => {
      const exportBtn = wrapper.find('[data-test="users-export-btn"]');
      await exportBtn.trigger("click");
      await flushPromises();

      const dialog = getDialog();
      const numberInput = dialog.find('[data-test="number-of-namespaces-input"] input');
      await numberInput.setValue("10");
      await flushPromises();

      expect(dialog.text()).not.toContain("must be greater than or equal to 0");
    });

    it("disables export button when validation fails", async () => {
      const exportBtn = wrapper.find('[data-test="users-export-btn"]');
      await exportBtn.trigger("click");
      await flushPromises();

      const dialog = getDialog();
      const numberInput = dialog.find('[data-test="number-of-namespaces-input"] input');
      await numberInput.setValue("-1");
      await flushPromises();

      const confirmBtn = dialog.find('[data-test="confirm-btn"]');
      expect(confirmBtn.attributes("disabled")).toBeDefined();
    });
  });

  describe("exporting users", () => {
    it("exports with 'more than' filter", async () => {
      mountWrapper();
      const exportBtn = wrapper.find('[data-test="users-export-btn"]');
      await exportBtn.trigger("click");
      await flushPromises();

      const dialog = getDialog();
      const numberInput = dialog.find('[data-test="number-of-namespaces-input"] input');
      await numberInput.setValue("5");
      await flushPromises();

      const confirmBtn = dialog.find('[data-test="confirm-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(usersStore.exportUsersToCsv).toHaveBeenCalledWith(expect.any(String));
      expect(saveAs).toHaveBeenCalledWith(
        expect.any(Blob),
        "users_with_more_than_5_namespaces.csv",
      );
      expect(mockSnackbar.showSuccess).toHaveBeenCalledWith("Exported users successfully.");
    });

    it("exports with 'exactly' filter", async () => {
      mountWrapper();
      const exportBtn = wrapper.find('[data-test="users-export-btn"]');
      await exportBtn.trigger("click");
      await flushPromises();

      const dialog = getDialog();
      const exactlyRadio = dialog.find('[data-test="radio-exactly"] input');
      await exactlyRadio.setValue(true);
      await flushPromises();

      const numberInput = dialog.find('[data-test="number-of-namespaces-input"] input');
      await numberInput.setValue("3");
      await flushPromises();

      const confirmBtn = dialog.find('[data-test="confirm-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(usersStore.exportUsersToCsv).toHaveBeenCalled();
      expect(saveAs).toHaveBeenCalledWith(
        expect.any(Blob),
        "users_with_exactly_3_namespaces.csv",
      );
      expect(mockSnackbar.showSuccess).toHaveBeenCalledWith("Exported users successfully.");
    });

    it("exports with zero namespaces", async () => {
      mountWrapper();
      const exportBtn = wrapper.find('[data-test="users-export-btn"]');
      await exportBtn.trigger("click");
      await flushPromises();

      const dialog = getDialog();
      const confirmBtn = dialog.find('[data-test="confirm-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(usersStore.exportUsersToCsv).toHaveBeenCalled();
      expect(saveAs).toHaveBeenCalledWith(
        expect.any(Blob),
        "users_with_more_than_0_namespaces.csv",
      );
      expect(mockSnackbar.showSuccess).toHaveBeenCalledWith("Exported users successfully.");
    });

    it("closes dialog after successful export", async () => {
      mountWrapper();
      const exportBtn = wrapper.find('[data-test="users-export-btn"]');
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
      vi.mocked(usersStore.exportUsersToCsv).mockRejectedValueOnce(
        createAxiosError(500, "Internal Server Error"),
      );

      const exportBtn = wrapper.find('[data-test="users-export-btn"]');
      await exportBtn.trigger("click");
      await flushPromises();

      const dialog = getDialog();
      const confirmBtn = dialog.find('[data-test="confirm-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to export users.");
      expect(saveAs).not.toHaveBeenCalled();
    });

    it("keeps dialog open when export fails", async () => {
      mountWrapper();
      vi.mocked(usersStore.exportUsersToCsv).mockRejectedValueOnce(
        createAxiosError(500, "Internal Server Error"),
      );

      const exportBtn = wrapper.find('[data-test="users-export-btn"]');
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
      const exportBtn = wrapper.find('[data-test="users-export-btn"]');
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
      const exportBtn = wrapper.find('[data-test="users-export-btn"]');
      await exportBtn.trigger("click");
      await flushPromises();

      let dialog = getDialog();
      const numberInput = dialog.find('[data-test="number-of-namespaces-input"] input');
      await numberInput.setValue("42");
      await flushPromises();

      const cancelBtn = dialog.find('[data-test="cancel-btn"]');
      await cancelBtn.trigger("click");
      await flushPromises();

      // Reopen dialog
      await exportBtn.trigger("click");
      await flushPromises();

      dialog = getDialog();
      const numberInputReopened = dialog.find('[data-test="number-of-namespaces-input"] input');
      expect((numberInputReopened.element as HTMLInputElement).value).toBe("0");
    });

    it("resets filter selection when dialog is closed and reopened", async () => {
      const exportBtn = wrapper.find('[data-test="users-export-btn"]');
      await exportBtn.trigger("click");
      await flushPromises();

      let dialog = getDialog();
      const exactlyRadio = dialog.find('[data-test="radio-exactly"] input');
      await exactlyRadio.setValue(true);
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
