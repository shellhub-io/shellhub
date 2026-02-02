import { describe, expect, it, afterEach, beforeEach, vi } from "vitest";
import { VueWrapper, DOMWrapper, flushPromises } from "@vue/test-utils";
import { mockSnackbar, mountComponent } from "@tests/utils/mount";
import PrivateKeyDelete from "@/components/PrivateKeys/PrivateKeyDelete.vue";
import usePrivateKeysStore from "@/store/modules/private_keys";

describe("PrivateKeyDelete", () => {
  let wrapper: VueWrapper<InstanceType<typeof PrivateKeyDelete>>;
  let dialog: DOMWrapper<Element>;
  let privateKeysStore: ReturnType<typeof usePrivateKeysStore>;

  const openDialog = async () => {
    await wrapper.find('[data-test="private-key-delete-btn"]').trigger("click");
    await flushPromises();
  };

  beforeEach(() => {
    wrapper = mountComponent(PrivateKeyDelete, {
      props: { id: 1 },
      attachTo: document.body,
    });

    privateKeysStore = usePrivateKeysStore();
    dialog = new DOMWrapper(document.body);
  });

  afterEach(() => {
    wrapper?.unmount();
    vi.clearAllMocks();
    document.body.innerHTML = "";
  });

  describe("Delete button", () => {
    it("Renders delete button", () => {
      expect(wrapper.find('[data-test="private-key-delete-btn"]').exists()).toBe(true);
    });

    it("Shows delete icon", () => {
      const icon = wrapper.find('[data-test="private-key-delete-icon"]');
      expect(icon.exists()).toBe(true);
      expect(icon.find(".v-icon").classes()).toContain("mdi-delete");
    });

    it("Shows Remove text", () => {
      const title = wrapper.find('[data-test="private-key-delete-btn-title"]');
      expect(title.text()).toBe("Remove");
    });

    it("Opens dialog when clicked", async () => {
      await openDialog();

      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(messageDialog.props("modelValue")).toBe(true);
    });
  });

  describe("Delete dialog", () => {
    beforeEach(() => openDialog());

    it("Shows MessageDialog with correct props", () => {
      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(messageDialog.props("title")).toBe("Are you sure?");
      expect(messageDialog.props("description")).toBe("You are about to delete this private key");
      expect(messageDialog.props("icon")).toBe("mdi-alert");
      expect(messageDialog.props("iconColor")).toBe("error");
    });

    it("Shows Delete and Close buttons", () => {
      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(messageDialog.props("confirmText")).toBe("Delete");
      expect(messageDialog.props("confirmColor")).toBe("error");
      expect(messageDialog.props("cancelText")).toBe("Close");
    });

    it("Closes dialog when cancel is clicked", async () => {
      await dialog.find('[data-test="close-btn"]').trigger("click");
      await flushPromises();

      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(messageDialog.props("modelValue")).toBe(false);
    });
  });

  describe("Delete private key", () => {
    it("Calls deletePrivateKey when confirmed", async () => {
      await openDialog();

      await dialog.find('[data-test="confirm-btn"]').trigger("click");
      await flushPromises();

      expect(privateKeysStore.deletePrivateKey).toHaveBeenCalledWith(1);
    });

    it("Shows success message after deletion", async () => {
      await openDialog();

      await dialog.find('[data-test="confirm-btn"]').trigger("click");
      await flushPromises();

      expect(mockSnackbar.showSuccess).toHaveBeenCalledWith("The private key was removed successfully");
    });

    it("Emits update event after deletion", async () => {
      await openDialog();

      await dialog.find('[data-test="confirm-btn"]').trigger("click");
      await flushPromises();

      expect(wrapper.emitted("update")).toBeTruthy();
    });

    it("Closes dialog after deletion", async () => {
      await openDialog();

      await dialog.find('[data-test="confirm-btn"]').trigger("click");
      await flushPromises();

      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(messageDialog.props("modelValue")).toBe(false);
    });
  });
});
