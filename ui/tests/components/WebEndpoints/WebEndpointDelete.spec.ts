import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { flushPromises, VueWrapper } from "@vue/test-utils";
import WebEndpointDelete from "@/components/WebEndpoints/WebEndpointDelete.vue";
import { mountComponent, mockSnackbar } from "@tests/utils/mount";
import { createAxiosError } from "@tests/utils/axiosError";
import useWebEndpointsStore from "@/store/modules/web_endpoints";

describe("WebEndpointDelete", () => {
  let wrapper: VueWrapper<InstanceType<typeof WebEndpointDelete>>;
  let webEndpointsStore: ReturnType<typeof useWebEndpointsStore>;

  const mountWrapper = ({ address = "test-address-123", role = "owner" } = { }) => {
    wrapper = mountComponent(WebEndpointDelete, {
      props: { address },
      piniaOptions: { initialState: { auth: { role } } },
    });

    webEndpointsStore = useWebEndpointsStore();
  };

  beforeEach(() => mountWrapper());

  afterEach(() => {
    wrapper?.unmount();
    vi.clearAllMocks();
  });

  describe("Component rendering", () => {
    it("renders delete button", () => {
      const deleteBtn = wrapper.find('[data-test="web-endpoint-delete-dialog-btn"]');
      expect(deleteBtn.exists()).toBe(true);
      expect(deleteBtn.find("i").classes()).toContain("mdi-delete");
    });

    it("disables button when user lacks permission", () => {
      wrapper.unmount();
      mountWrapper({ role: "observer" });

      const deleteBtn = wrapper.find('[data-test="web-endpoint-delete-dialog-btn"]');
      expect(deleteBtn.attributes("disabled")).toBe("");
    });

    it("enables button when user has permission", () => {
      const deleteBtn = wrapper.find('[data-test="web-endpoint-delete-dialog-btn"]');
      expect(deleteBtn.attributes("disabled")).toBeUndefined();
    });
  });

  describe("Dialog behavior", () => {
    it("does not show dialog initially", () => {
      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(messageDialog.props("modelValue")).toBe(false);
    });

    it("opens dialog when delete button is clicked", async () => {
      const deleteBtn = wrapper.find('[data-test="web-endpoint-delete-dialog-btn"]');
      await deleteBtn.trigger("click");
      await flushPromises();

      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(messageDialog.props("modelValue")).toBe(true);
    });

    it("renders MessageDialog with correct props", async () => {
      const deleteBtn = wrapper.find('[data-test="web-endpoint-delete-dialog-btn"]');
      await deleteBtn.trigger("click");
      await flushPromises();

      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(messageDialog.props("title")).toBe("Are you sure?");
      expect(messageDialog.props("description")).toBe("You are about to remove this Web Endpoint.");
      expect(messageDialog.props("icon")).toBe("mdi-alert");
      expect(messageDialog.props("iconColor")).toBe("error");
      expect(messageDialog.props("confirmText")).toBe("Delete Web Endpoint");
      expect(messageDialog.props("confirmColor")).toBe("error");
      expect(messageDialog.props("cancelText")).toBe("Close");
    });

    it("closes dialog when cancel is clicked", async () => {
      const deleteBtn = wrapper.find('[data-test="web-endpoint-delete-dialog-btn"]');
      await deleteBtn.trigger("click");
      await flushPromises();

      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      messageDialog.vm.$emit("cancel");
      await flushPromises();

      expect(messageDialog.props("modelValue")).toBe(false);
    });

    it("closes dialog when close is clicked", async () => {
      const deleteBtn = wrapper.find('[data-test="web-endpoint-delete-dialog-btn"]');
      await deleteBtn.trigger("click");
      await flushPromises();

      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      messageDialog.vm.$emit("close");
      await flushPromises();

      expect(messageDialog.props("modelValue")).toBe(false);
    });
  });

  describe("Web endpoint deletion", () => {
    it("deletes web endpoint successfully", async () => {
      const deleteBtn = wrapper.find('[data-test="web-endpoint-delete-dialog-btn"]');
      await deleteBtn.trigger("click");
      await flushPromises();

      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      messageDialog.vm.$emit("confirm");
      await flushPromises();

      expect(webEndpointsStore.deleteWebEndpoint).toHaveBeenCalledWith("test-address-123");
      expect(mockSnackbar.showSuccess).toHaveBeenCalledWith("Web Endpoint deleted successfully.");
      expect(wrapper.emitted("update")).toBeTruthy();
      expect(messageDialog.props("modelValue")).toBe(false);
    });

    it("deletes web endpoint with different address", async () => {
      wrapper.unmount();
      mountWrapper({ address: "another-address-456" });

      const deleteBtn = wrapper.find('[data-test="web-endpoint-delete-dialog-btn"]');
      await deleteBtn.trigger("click");
      await flushPromises();

      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      messageDialog.vm.$emit("confirm");
      await flushPromises();

      expect(webEndpointsStore.deleteWebEndpoint).toHaveBeenCalledWith("another-address-456");
    });
  });

  describe("Error handling", () => {
    it("shows error when deletion fails", async () => {
      const error = createAxiosError(500, "Internal Server Error");
      vi.mocked(webEndpointsStore.deleteWebEndpoint).mockRejectedValueOnce(error);

      const deleteBtn = wrapper.find('[data-test="web-endpoint-delete-dialog-btn"]');
      await deleteBtn.trigger("click");
      await flushPromises();

      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      messageDialog.vm.$emit("confirm");
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to delete Web Endpoint.");
    });
  });
});
