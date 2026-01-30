import { describe, expect, it, afterEach, vi, beforeEach } from "vitest";
import { VueWrapper, DOMWrapper, flushPromises } from "@vue/test-utils";
import { mockSnackbar, mountComponent } from "@tests/utils/mount";
import { createAxiosError } from "@tests/utils/axiosError";
import { mockNamespace } from "@tests/mocks";
import ConnectionAnnouncementEdit from "@/components/Namespace/ConnectionAnnouncementEdit.vue";
import useNamespacesStore from "@/store/modules/namespaces";
import handleError from "@/utils/handleError";

const mockNamespaceWithAnnouncement = {
  ...mockNamespace,
  settings: {
    connection_announcement: "Welcome message",
  },
};

describe("ConnectionAnnouncementEdit", () => {
  let wrapper: VueWrapper<InstanceType<typeof ConnectionAnnouncementEdit>>;
  let dialog: DOMWrapper<Element>;
  let namespacesStore: ReturnType<typeof useNamespacesStore>;

  const mountWrapper = async (currentNamespace = mockNamespaceWithAnnouncement) => {
    wrapper = mountComponent(ConnectionAnnouncementEdit, {
      props: { modelValue: true },
      attachTo: document.body,
      piniaOptions: {
        initialState: {
          namespaces: { currentNamespace },
          auth: { token: "test-token", tenantId: currentNamespace.tenant_id },
        },
      },
    });

    namespacesStore = useNamespacesStore();
    dialog = new DOMWrapper(document.body);

    await flushPromises();
  };

  afterEach(() => {
    wrapper?.unmount();
    vi.clearAllMocks();
    document.body.innerHTML = "";
  });

  describe("Dialog display", () => {
    beforeEach(() => mountWrapper());
    it("Renders FormDialog component", () => {
      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      expect(formDialog.exists()).toBe(true);
    });

    it("Shows correct title", () => {
      expect(dialog.find('[data-test="window-dialog-titlebar"]').text()).toContain("Change Connection Announcement");
    });

    it("Shows Edit and Close buttons", () => {
      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      expect(formDialog.props("confirmText")).toBe("Save Announcement");
      expect(formDialog.props("cancelText")).toBe("Cancel");
    });
  });

  describe("Textarea field", () => {
    it("Renders textarea with current announcement", async () => {
      await mountWrapper();

      const textarea = dialog.find('[data-test="connection-announcement-text"] textarea');

      expect(textarea.exists()).toBe(true);
      expect((textarea.element as HTMLTextAreaElement).value).toBe("Welcome message");
    });

    it("Renders empty textarea when no announcement exists", async () => {
      const currentNamespace = {
        ...mockNamespace,
        settings: {
          connection_announcement: "",
        },
      };
      await mountWrapper(currentNamespace);

      const textarea = dialog.find('[data-test="connection-announcement-text"] textarea');
      expect((textarea.element as HTMLTextAreaElement).value).toBe("");
    });
  });

  describe("Form validation", () => {
    beforeEach(() => mountWrapper());
    it("Validates max length of 4096 characters", async () => {
      const textarea = dialog.find('[data-test="connection-announcement-text"] textarea');
      const longText = "a".repeat(4097);

      await textarea.setValue(longText);
      await flushPromises();

      const confirmBtn = dialog.find('[data-test="change-connection-announcement-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      const errorMessage = dialog.find('[data-test="connection-announcement-text"] .v-messages__message');
      expect(errorMessage.text()).toBe("Your message should be 1-4096 characters long");
    });

    it("Accepts text within 4096 character limit", async () => {
      const textarea = dialog.find('[data-test="connection-announcement-text"] textarea');
      const validText = "a".repeat(4096);

      await textarea.setValue(validText);
      await flushPromises();

      const errorMessage = dialog.find('[data-test="connection-announcement-text"] .v-messages__message');
      expect(errorMessage.text()).not.toBe("Your message should be 1-4096 characters long");
    });
  });

  describe("Update announcement", () => {
    beforeEach(() => mountWrapper());
    it("Calls editNamespace when confirmed", async () => {
      const textarea = dialog.find('[data-test="connection-announcement-text"] textarea');
      const newAnnouncement = "Updated welcome message";
      await textarea.setValue(newAnnouncement);
      await flushPromises();

      const confirmBtn = dialog.find('[data-test="change-connection-announcement-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(namespacesStore.editNamespace).toHaveBeenCalledWith({
        tenant_id: mockNamespace.tenant_id,
        settings: { connection_announcement: newAnnouncement },
      });
    });

    it("Emits update event after successful edit", async () => {
      const textarea = dialog.find('[data-test="connection-announcement-text"] textarea');
      await textarea.setValue("Updated message");
      await flushPromises();

      const confirmBtn = dialog.find('[data-test="change-connection-announcement-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(wrapper.emitted("update")).toBeTruthy();
    });
  });

  describe("Error handling", () => {
    beforeEach(() => mountWrapper());
    it("Handles 400 error", async () => {
      const error = createAxiosError(400, "Bad Request");

      vi.mocked(namespacesStore.editNamespace).mockRejectedValueOnce(error);

      const confirmBtn = dialog.find('[data-test="change-connection-announcement-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(namespacesStore.editNamespace).toHaveBeenCalled();
      const errorMessage = dialog.find('[data-test="connection-announcement-text"] .v-messages__message');
      expect(errorMessage.text()).toBe("This message is not valid");
      expect(handleError).not.toHaveBeenCalled();
    });

    it("Handles 500 error", async () => {
      const error = createAxiosError(500, "Internal Server Error");

      vi.mocked(namespacesStore.editNamespace).mockRejectedValueOnce(error);

      const confirmBtn = dialog.find('[data-test="change-connection-announcement-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(namespacesStore.editNamespace).toHaveBeenCalled();
      expect(mockSnackbar.showError).toHaveBeenCalledWith("An error occurred while updating the connection announcement.");
      expect(handleError).toHaveBeenCalledWith(error);
    });
  });
});
