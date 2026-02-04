import { describe, expect, it, beforeEach, afterEach, vi } from "vitest";
import { VueWrapper, flushPromises } from "@vue/test-utils";
import { mockSnackbar, mountComponent } from "@tests/utils/mount";
import { createAxiosError } from "@tests/utils/axiosError";
import SessionDelete from "@/components/Sessions/SessionDelete.vue";
import useSessionsStore from "@/store/modules/sessions";
import handleError from "@/utils/handleError";

describe("SessionDelete", () => {
  let wrapper: VueWrapper<InstanceType<typeof SessionDelete>>;
  let sessionsStore: ReturnType<typeof useSessionsStore>;

  const mountWrapper = (hasAuthorization = true) => {
    wrapper = mountComponent(SessionDelete, {
      props: {
        uid: "session-123",
        hasAuthorization,
      },
    });

    sessionsStore = useSessionsStore();
  };

  beforeEach(() => mountWrapper());

  afterEach(() => {
    wrapper?.unmount();
    vi.clearAllMocks();
  });

  describe("List item rendering", () => {
    it("Renders list item", () => {
      const listItem = wrapper.find(".v-list-item");
      expect(listItem.exists()).toBe(true);
    });

    it("Shows delete session icon", () => {
      const icon = wrapper.find(".v-icon");
      expect(icon.exists()).toBe(true);
      expect(icon.classes()).toContain("mdi-playlist-remove");
    });

    it("Shows delete session text", () => {
      const title = wrapper.find('[data-test="mdi-information-list-item"]');
      expect(title.text()).toBe("Delete Session Record");
    });

    it("Disables list item when no authorization", () => {
      wrapper.unmount();
      mountWrapper(false);

      const listItem = wrapper.find(".v-list-item");
      expect(listItem.classes()).toContain("v-list-item--disabled");
    });

    it("Enables list item when authorized", () => {
      const listItem = wrapper.find(".v-list-item");
      expect(listItem.classes()).not.toContain("v-list-item--disabled");
    });
  });

  describe("Dialog interaction", () => {
    it("Opens dialog when list item is clicked", async () => {
      const listItem = wrapper.find(".v-list-item");
      await listItem.trigger("click");

      const dialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(dialog.props("modelValue")).toBe(true);
    });

    it("Dialog has correct title", () => {
      const dialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(dialog.props("title")).toBe("Are you sure?");
    });

    it("Dialog has correct description", () => {
      const dialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(dialog.props("description")).toBe(
        "You are going to delete the logs recorded for this session. After confirming this action cannot be undone.",
      );
    });

    it("Dialog has correct icon", () => {
      const dialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(dialog.props("icon")).toBe("mdi-playlist-remove");
      expect(dialog.props("iconColor")).toBe("error");
    });

    it("Dialog has correct button texts", () => {
      const dialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(dialog.props("confirmText")).toBe("Remove");
      expect(dialog.props("confirmColor")).toBe("error");
      expect(dialog.props("cancelText")).toBe("Close");
    });

    it("Closes dialog when cancel is clicked", async () => {
      const listItem = wrapper.find(".v-list-item");
      await listItem.trigger("click");

      const dialog = wrapper.findComponent({ name: "MessageDialog" });
      await dialog.vm.$emit("cancel");

      expect(dialog.props("modelValue")).toBe(false);
    });

    it("Closes dialog on close event", async () => {
      const listItem = wrapper.find(".v-list-item");
      await listItem.trigger("click");

      const dialog = wrapper.findComponent({ name: "MessageDialog" });
      await dialog.vm.$emit("close");

      expect(dialog.props("modelValue")).toBe(false);
    });
  });

  describe("Delete session action", () => {
    it("Calls deleteSessionLogs with correct parameter", async () => {
      const listItem = wrapper.find(".v-list-item");
      await listItem.trigger("click");

      const dialog = wrapper.findComponent({ name: "MessageDialog" });
      await dialog.vm.$emit("confirm");
      await flushPromises();

      expect(sessionsStore.deleteSessionLogs).toHaveBeenCalledWith("session-123");
    });

    it("Shows success message on successful delete", async () => {
      const listItem = wrapper.find(".v-list-item");
      await listItem.trigger("click");

      const dialog = wrapper.findComponent({ name: "MessageDialog" });
      await dialog.vm.$emit("confirm");
      await flushPromises();

      expect(mockSnackbar.showSuccess).toHaveBeenCalledWith(
        "Successfully deleted the session logs.",
      );
    });

    it("Closes dialog after successful delete", async () => {
      const listItem = wrapper.find(".v-list-item");
      await listItem.trigger("click");

      const dialog = wrapper.findComponent({ name: "MessageDialog" });
      await dialog.vm.$emit("confirm");
      await flushPromises();

      expect(dialog.props("modelValue")).toBe(false);
    });

    it("Emits update event after successful delete", async () => {
      const listItem = wrapper.find(".v-list-item");
      await listItem.trigger("click");

      const dialog = wrapper.findComponent({ name: "MessageDialog" });
      await dialog.vm.$emit("confirm");
      await flushPromises();

      expect(wrapper.emitted("update")).toBeTruthy();
    });
  });

  describe("Error handling", () => {
    it("Shows error message on failed delete", async () => {
      const error = createAxiosError(500, "Internal Server Error");
      vi.mocked(sessionsStore.deleteSessionLogs).mockRejectedValueOnce(error);

      const listItem = wrapper.find(".v-list-item");
      await listItem.trigger("click");

      const dialog = wrapper.findComponent({ name: "MessageDialog" });
      await dialog.vm.$emit("confirm");
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith(
        "An error occurred while deleting the session logs.",
      );
      expect(handleError).toHaveBeenCalledWith(error);
    });
  });
});
