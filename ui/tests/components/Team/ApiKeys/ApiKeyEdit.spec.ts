import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { DOMWrapper, VueWrapper, flushPromises } from "@vue/test-utils";
import { createAxiosError } from "@tests/utils/axiosError";
import { mountComponent, mockSnackbar } from "@tests/utils/mount";
import ApiKeyEdit from "@/components/Team/ApiKeys/ApiKeyEdit.vue";
import useApiKeysStore from "@/store/modules/api_keys";
import { BasicRole } from "@/interfaces/INamespace";

describe("ApiKeyEdit", () => {
  let wrapper: VueWrapper<InstanceType<typeof ApiKeyEdit>>;
  let apiKeysStore: ReturnType<typeof useApiKeysStore>;
  let dialog: DOMWrapper<HTMLElement>;

  const openDialog = async () => {
    const listItem = wrapper.find('[data-test="edit-api-key-btn"]');
    await listItem.trigger("click");
    await flushPromises();
  };

  const mountWrapper = ({
    keyName = "test-key",
    keyRole = "administrator" as BasicRole,
    hasAuthorization = true,
    disabled = false,
  } = {}) => {
    wrapper = mountComponent(ApiKeyEdit, {
      props: { keyName, keyRole, hasAuthorization, disabled },
      attachTo: document.body,
    });
    apiKeysStore = useApiKeysStore();
    dialog = new DOMWrapper(document.body);
  };

  beforeEach(() => mountWrapper());

  afterEach(() => {
    wrapper?.unmount();
    vi.clearAllMocks();
    document.body.innerHTML = "";
  });

  describe("Rendering", () => {
    it("renders edit list item", () => {
      const listItem = wrapper.find('[data-test="edit-main-btn-title"]');
      expect(listItem.exists()).toBe(true);
      expect(listItem.text()).toBe("Edit");
    });

    it("renders edit icon", () => {
      const icon = wrapper.find('[data-test="edit-icon"]');
      expect(icon.exists()).toBe(true);
    });

    it("disables list item when hasAuthorization is false", () => {
      wrapper.unmount();
      mountWrapper({ hasAuthorization: false });

      const listItem = wrapper.find("[data-test=edit-api-key-btn]");
      expect(listItem.classes()).toContain("v-list-item--disabled");
    });

    it("disables list item when disabled prop is true", () => {
      wrapper.unmount();
      mountWrapper({ disabled: true });

      const listItem = wrapper.find("[data-test=edit-api-key-btn]");
      expect(listItem.classes()).toContain("v-list-item--disabled");
    });
  });

  describe("Dialog", () => {
    it("opens dialog when list item is clicked", async () => {
      await openDialog();

      const editDialog = dialog.find('[data-test="edit-dialog"]');
      expect(editDialog.exists()).toBe(true);
    });

    it("renders form fields with initial values", async () => {
      wrapper.unmount();
      mountWrapper({ keyName: "my-key", keyRole: "observer" });
      await openDialog();

      const nameInput = dialog.find('[data-test="key-name-text"] input').element as HTMLInputElement;
      expect(nameInput.value).toBe("my-key");
    });

    it("renders dialog buttons", async () => {
      await openDialog();

      expect(dialog.find('[data-test="edit-btn"]').exists()).toBe(true);
      expect(dialog.find('[data-test="close-btn"]').exists()).toBe(true);
    });
  });

  describe("API key editing", () => {
    it("calls editApiKey when submitting with changed name", async () => {
      await openDialog();

      const nameInput = dialog.find('[data-test="key-name-text"] input');
      await nameInput.setValue("updated-key");
      await flushPromises();

      const saveBtn = dialog.find('[data-test="edit-btn"]');
      await saveBtn.trigger("click");
      await flushPromises();

      expect(apiKeysStore.editApiKey).toHaveBeenCalledWith({
        key: "test-key",
        name: "updated-key",
        role: "administrator",
      });
    });

    it("does not send name when it hasn't changed", async () => {
      await openDialog();

      const saveBtn = dialog.find('[data-test="edit-btn"]');
      await saveBtn.trigger("click");
      await flushPromises();

      expect(apiKeysStore.editApiKey).toHaveBeenCalledWith({
        key: "test-key",
        name: undefined,
        role: "administrator",
      });
    });

    it("shows success snackbar on successful edit", async () => {
      await openDialog();

      const nameInput = dialog.find('[data-test="key-name-text"] input');
      await nameInput.setValue("updated-key");
      await flushPromises();

      const saveBtn = dialog.find('[data-test="edit-btn"]');
      await saveBtn.trigger("click");
      await flushPromises();

      expect(mockSnackbar.showSuccess).toHaveBeenCalledWith("API Key edited successfully.");
    });

    it("emits update event on successful edit", async () => {
      await openDialog();

      const nameInput = dialog.find('[data-test="key-name-text"] input');
      await nameInput.setValue("updated-key");
      await flushPromises();

      const saveBtn = dialog.find('[data-test="edit-btn"]');
      await saveBtn.trigger("click");
      await flushPromises();

      expect(wrapper.emitted("update")).toBeTruthy();
    });

    it("closes dialog on successful edit", async () => {
      await openDialog();

      const nameInput = dialog.find('[data-test="key-name-text"] input');
      await nameInput.setValue("updated-key");
      await flushPromises();

      const saveBtn = dialog.find('[data-test="edit-btn"]');
      await saveBtn.trigger("click");
      await flushPromises();

      const dialogContent = dialog.find(".v-overlay__content");
      expect(dialogContent.attributes("style")).toContain("display: none;");
    });
  });

  describe("Error handling", () => {
    it("shows error snackbar when edit fails", async () => {
      const error = createAxiosError(500, "Internal Server Error");
      vi.mocked(apiKeysStore.editApiKey).mockRejectedValueOnce(error);

      await openDialog();

      const nameInput = dialog.find('[data-test="key-name-text"] input');
      await nameInput.setValue("updated-key");
      await flushPromises();

      const saveBtn = dialog.find('[data-test="edit-btn"]');
      await saveBtn.trigger("click");
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to edit API Key.");
    });

    it("shows error message for 409 status", async () => {
      const error = createAxiosError(409, "Conflict");
      vi.mocked(apiKeysStore.editApiKey).mockRejectedValueOnce(error);

      await openDialog();

      const nameInput = dialog.find('[data-test="key-name-text"] input');
      await nameInput.setValue("updated-key");
      await flushPromises();

      const saveBtn = dialog.find('[data-test="edit-btn"]');
      await saveBtn.trigger("click");
      await flushPromises();

      const alert = dialog.find('[data-test="form-dialog-alert"]');
      expect(alert.text()).toContain("An API key with the same name already exists.");
    });

    it("shows generic error message for other status codes", async () => {
      const error = createAxiosError(503, "Service Unavailable");
      vi.mocked(apiKeysStore.editApiKey).mockRejectedValueOnce(error);

      await openDialog();

      const nameInput = dialog.find('[data-test="key-name-text"] input');
      await nameInput.setValue("updated-key");
      await flushPromises();

      const saveBtn = dialog.find('[data-test="edit-btn"]');
      await saveBtn.trigger("click");
      await flushPromises();

      const alert = dialog.find('[data-test="form-dialog-alert"]');
      expect(alert.text()).toContain("An error occurred while editing your API key.");
    });

    it("dismisses error message when alert is dismissed", async () => {
      const error = createAxiosError(409, "Conflict");
      vi.mocked(apiKeysStore.editApiKey).mockRejectedValueOnce(error);

      await openDialog();

      const nameInput = dialog.find('[data-test="key-name-text"] input');
      await nameInput.setValue("updated-key");
      await flushPromises();

      const saveBtn = dialog.find('[data-test="edit-btn"]');
      await saveBtn.trigger("click");
      await flushPromises();

      let alert = dialog.find('[data-test="form-dialog-alert"]');
      expect(alert.exists()).toBe(true);

      const closeAlertBtn = alert.find('[data-test="alert-got-it-btn"]');
      await closeAlertBtn.trigger("click");
      await flushPromises();

      alert = dialog.find('[data-test="form-dialog-alert"]');
      expect(alert.exists()).toBe(false);
    });
  });

  describe("Dialog close", () => {
    it("closes dialog when cancel button is clicked", async () => {
      await openDialog();

      const cancelBtn = dialog.find('[data-test="close-btn"]');
      await cancelBtn.trigger("click");
      await flushPromises();

      const dialogContent = dialog.find(".v-overlay__content");
      expect(dialogContent.attributes("style")).toContain("display: none;");
    });
  });
});
