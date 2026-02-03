import { describe, expect, it, afterEach, beforeEach, vi } from "vitest";
import { VueWrapper, DOMWrapper, flushPromises } from "@vue/test-utils";
import { mockSnackbar, mountComponent } from "@tests/utils/mount";
import PublicKeyDelete from "@/components/PublicKeys/PublicKeyDelete.vue";
import usePublicKeysStore from "@/store/modules/public_keys";
import handleError from "@/utils/handleError";
import { createAxiosError } from "@tests/utils/axiosError";

describe("PublicKeyDelete", () => {
  let wrapper: VueWrapper<InstanceType<typeof PublicKeyDelete>>;
  let dialog: DOMWrapper<Element>;
  let publicKeysStore: ReturnType<typeof usePublicKeysStore>;

  const openDialog = async () => {
    const deleteBtn = wrapper.find('[data-test="public-key-remove-btn"]');
    await deleteBtn.trigger("click");
    await flushPromises();
  };

  const mountWrapper = (hasAuthorization = true) => {
    wrapper = mountComponent(PublicKeyDelete, {
      props: {
        fingerprint: "aa:bb:cc:dd",
        hasAuthorization,
      },
      attachTo: document.body,
    });

    publicKeysStore = usePublicKeysStore();
    dialog = new DOMWrapper(document.body);
  };

  beforeEach(() => mountWrapper());

  afterEach(() => {
    wrapper?.unmount();
    vi.clearAllMocks();
    document.body.innerHTML = "";
  });

  describe("Delete button", () => {
    it("Renders delete button", () => {
      const deleteBtn = wrapper.find('[data-test="public-key-remove-btn"]');
      expect(deleteBtn.exists()).toBe(true);
    });

    it("Shows delete icon", () => {
      const deleteBtn = wrapper.find('[data-test="public-key-remove-btn"]');
      const icon = deleteBtn.find('[data-test="remove-icon"]');
      expect(icon.exists()).toBe(true);
      expect(icon.classes()).toContain("mdi-delete");
    });

    it("Shows 'Remove' text", () => {
      const deleteBtn = wrapper.find('[data-test="public-key-remove-btn"]');
      const title = deleteBtn.find('[data-test="remove-title"]');
      expect(title.text()).toBe("Remove");
    });

    it("Is disabled when hasAuthorization is false", () => {
      wrapper.unmount();
      mountWrapper(false);

      const deleteBtn = wrapper.find('[data-test="public-key-remove-btn"]');
      expect(deleteBtn.classes()).toContain("v-list-item--disabled");
    });

    it("Opens dialog when clicked", async () => {
      await openDialog();

      const messageDialog = dialog.find('[data-test="delete-public-key-dialog"]');
      expect(messageDialog.exists()).toBe(true);
    });
  });

  describe("Dialog display", () => {
    beforeEach(() => openDialog());

    it("Shows MessageDialog with correct props", () => {
      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(messageDialog.props("title")).toBe("Are you sure?");
      expect(messageDialog.props("description")).toBe("You are about to delete this public key");
      expect(messageDialog.props("icon")).toBe("mdi-alert");
      expect(messageDialog.props("iconColor")).toBe("error");
      expect(messageDialog.props("confirmText")).toBe("Delete");
      expect(messageDialog.props("confirmColor")).toBe("error");
      expect(messageDialog.props("cancelText")).toBe("Close");
    });
  });

  describe("Delete confirmation", () => {
    beforeEach(() => openDialog());

    it("Calls deletePublicKey when confirm is clicked", async () => {
      const confirmBtn = dialog.find('[data-test="confirm-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(publicKeysStore.deletePublicKey).toHaveBeenCalledWith("aa:bb:cc:dd");
    });

    it("Shows success message after deletion", async () => {
      const confirmBtn = dialog.find('[data-test="confirm-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(mockSnackbar.showSuccess).toHaveBeenCalledWith("The public key was removed successfully");
    });

    it("Emits update event after deletion", async () => {
      const confirmBtn = dialog.find('[data-test="confirm-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(wrapper.emitted("update")).toBeTruthy();
    });

    it("Closes dialog after deletion", async () => {
      const confirmBtn = dialog.find('[data-test="confirm-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(messageDialog.props("modelValue")).toBe(false);
    });
  });

  describe("Error handling", () => {
    beforeEach(() => openDialog());

    it("Handles deletion error", async () => {
      const error = createAxiosError(500, "Internal Server Error");
      vi.mocked(publicKeysStore.deletePublicKey).mockRejectedValueOnce(error);

      const confirmBtn = dialog.find('[data-test="confirm-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to remove the public key.");
      expect(handleError).toHaveBeenCalledWith(error);
    });

    it("Closes dialog even when error occurs", async () => {
      const error = createAxiosError(500, "Internal Server Error");
      vi.mocked(publicKeysStore.deletePublicKey).mockRejectedValueOnce(error);

      const confirmBtn = dialog.find('[data-test="confirm-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(messageDialog.props("modelValue")).toBe(false);
    });
  });

  describe("Dialog actions", () => {
    beforeEach(() => openDialog());

    it("Closes dialog when cancel is clicked", async () => {
      const cancelBtn = dialog.find('[data-test="close-btn"]');
      await cancelBtn.trigger("click");
      await flushPromises();

      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(messageDialog.props("modelValue")).toBe(false);
    });

    it("Does not delete key when cancel is clicked", async () => {
      const cancelBtn = dialog.find('[data-test="close-btn"]');
      await cancelBtn.trigger("click");
      await flushPromises();

      expect(publicKeysStore.deletePublicKey).not.toHaveBeenCalled();
    });
  });
});
