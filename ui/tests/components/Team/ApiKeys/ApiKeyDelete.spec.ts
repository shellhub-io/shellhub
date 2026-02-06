import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { DOMWrapper, VueWrapper, flushPromises } from "@vue/test-utils";
import { createAxiosError } from "@tests/utils/axiosError";
import { mountComponent, mockSnackbar } from "@tests/utils/mount";
import ApiKeyDelete from "@/components/Team/ApiKeys/ApiKeyDelete.vue";
import useApiKeysStore from "@/store/modules/api_keys";
import handleError from "@/utils/handleError";

describe("ApiKeyDelete", () => {
  let wrapper: VueWrapper<InstanceType<typeof ApiKeyDelete>>;
  let apiKeysStore: ReturnType<typeof useApiKeysStore>;
  let dialog: DOMWrapper<HTMLElement>;

  const openDialog = async () => {
    const listItem = wrapper.find('[data-test="delete-api-key-btn"]');
    await listItem.trigger("click");
    await flushPromises();
  };

  const mountWrapper = ({
    keyId = "test-key",
    hasAuthorization = true,
  } = {}) => {
    wrapper = mountComponent(ApiKeyDelete, {
      props: { keyId, hasAuthorization },
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
    it("renders delete list item", () => {
      const listItem = wrapper.find('[data-test="delete-main-btn-title"]');
      expect(listItem.exists()).toBe(true);
      expect(listItem.text()).toBe("Delete");
    });

    it("renders delete icon", () => {
      const icon = wrapper.find('[data-test="delete-icon"]');
      expect(icon.exists()).toBe(true);
    });

    it("disables list item when hasAuthorization is false", () => {
      wrapper.unmount();
      mountWrapper({ hasAuthorization: false });

      const listItem = wrapper.find("[data-test=delete-api-key-btn]");
      expect(listItem.classes()).toContain("v-list-item--disabled");
    });
  });

  describe("Dialog", () => {
    it("renders dialog buttons", async () => {
      await openDialog();

      expect(dialog.find('[data-test="delete-btn"]').exists()).toBe(true);
      expect(dialog.find('[data-test="close-btn"]').exists()).toBe(true);
    });
  });

  describe("API key deletion", () => {
    it("calls removeApiKey when confirming deletion", async () => {
      await openDialog();

      const deleteBtn = dialog.find('[data-test="delete-btn"]');
      await deleteBtn.trigger("click");
      await flushPromises();

      expect(apiKeysStore.removeApiKey).toHaveBeenCalledWith({
        key: "test-key",
      });
    });

    it("shows success snackbar on successful deletion", async () => {
      await openDialog();

      const deleteBtn = dialog.find('[data-test="delete-btn"]');
      await deleteBtn.trigger("click");
      await flushPromises();

      expect(mockSnackbar.showSuccess).toHaveBeenCalledWith("Api Key deleted successfully.");
    });

    it("emits update event on successful deletion", async () => {
      await openDialog();

      const deleteBtn = dialog.find('[data-test="delete-btn"]');
      await deleteBtn.trigger("click");
      await flushPromises();

      expect(wrapper.emitted("update")).toBeTruthy();
    });

    it("closes dialog on successful deletion", async () => {
      await openDialog();

      const deleteBtn = dialog.find('[data-test="delete-btn"]');
      await deleteBtn.trigger("click");
      await flushPromises();

      const title = dialog.find(".v-card-title");
      expect(title.exists()).toBe(false);
    });
  });

  describe("Error handling", () => {
    it("shows error snackbar when deletion fails", async () => {
      const error = createAxiosError(500, "Internal Server Error");
      vi.mocked(apiKeysStore.removeApiKey).mockRejectedValueOnce(error);

      await openDialog();

      const deleteBtn = dialog.find('[data-test="delete-btn"]');
      await deleteBtn.trigger("click");
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to delete Api Key.");
      expect(handleError).toHaveBeenCalledWith(error);
    });

    it("keeps dialog open when deletion fails", async () => {
      const error = createAxiosError(500, "Internal Server Error");
      vi.mocked(apiKeysStore.removeApiKey).mockRejectedValueOnce(error);

      await openDialog();

      const deleteBtn = dialog.find('[data-test="delete-btn"]');
      await deleteBtn.trigger("click");
      await flushPromises();

      const dialogContent = dialog.find(".v-overlay__content");
      expect(dialogContent.attributes("style")).not.toContain("display: none;");
    });
  });

  describe("Dialog close", () => {
    it("closes dialog when cancel button is clicked", async () => {
      await openDialog();

      const cancelBtn = dialog.find('[data-test="close-btn"]');
      await cancelBtn.trigger("click");
      await flushPromises();

      const title = dialog.find(".v-card-title");
      expect(title.exists()).toBe(false);
    });
  });
});
