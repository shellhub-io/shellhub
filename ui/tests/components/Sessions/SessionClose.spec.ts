import { describe, expect, it, beforeEach, afterEach, vi } from "vitest";
import { VueWrapper, flushPromises } from "@vue/test-utils";
import { mockSnackbar, mountComponent } from "@tests/utils/mount";
import { mockDevice } from "@tests/mocks/device";
import { createAxiosError } from "@tests/utils/axiosError";
import SessionClose from "@/components/Sessions/SessionClose.vue";
import useSessionsStore from "@/store/modules/sessions";
import handleError from "@/utils/handleError";

describe("SessionClose", () => {
  let wrapper: VueWrapper<InstanceType<typeof SessionClose>>;
  let sessionsStore: ReturnType<typeof useSessionsStore>;

  const mountWrapper = (hasAuthorization = true) => {
    wrapper = mountComponent(SessionClose, {
      props: {
        uid: "session-123",
        device: mockDevice,
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

    it("Shows close session icon", () => {
      const icon = wrapper.find(".v-icon");
      expect(icon.exists()).toBe(true);
      expect(icon.classes()).toContain("mdi-close-circle");
    });

    it("Shows close session text", () => {
      const title = wrapper.find('[data-test="mdi-information-list-item"]');
      expect(title.text()).toBe("Close Session");
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
        "You are going to close connection for this device. After confirming this action cannot be undone.",
      );
    });

    it("Dialog has correct icon", () => {
      const dialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(dialog.props("icon")).toBe("mdi-close-circle");
      expect(dialog.props("iconColor")).toBe("error");
    });

    it("Dialog has correct button texts", () => {
      const dialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(dialog.props("confirmText")).toBe("Close");
      expect(dialog.props("confirmColor")).toBe("error");
      expect(dialog.props("cancelText")).toBe("Cancel");
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

  describe("Close session action", () => {
    it("Calls closeSession with correct parameters", async () => {
      const listItem = wrapper.find(".v-list-item");
      await listItem.trigger("click");

      const dialog = wrapper.findComponent({ name: "MessageDialog" });
      await dialog.vm.$emit("confirm");
      await flushPromises();

      expect(sessionsStore.closeSession).toHaveBeenCalledWith({
        uid: "session-123",
        device_uid: mockDevice.uid,
      });
    });

    it("Shows success message on successful close", async () => {
      const listItem = wrapper.find(".v-list-item");
      await listItem.trigger("click");

      const dialog = wrapper.findComponent({ name: "MessageDialog" });
      await dialog.vm.$emit("confirm");
      await flushPromises();

      expect(mockSnackbar.showSuccess).toHaveBeenCalledWith("Session closed successfully.");
    });

    it("Closes dialog after successful close", async () => {
      const listItem = wrapper.find(".v-list-item");
      await listItem.trigger("click");

      const dialog = wrapper.findComponent({ name: "MessageDialog" });
      await dialog.vm.$emit("confirm");
      await flushPromises();

      expect(dialog.props("modelValue")).toBe(false);
    });

    it("Emits update event after successful close", async () => {
      const listItem = wrapper.find(".v-list-item");
      await listItem.trigger("click");

      const dialog = wrapper.findComponent({ name: "MessageDialog" });
      await dialog.vm.$emit("confirm");
      await flushPromises();

      expect(wrapper.emitted("update")).toBeTruthy();
    });
  });

  describe("Error handling", () => {
    it("Shows error message on failed close", async () => {
      const error = createAxiosError(500, "Internal Server Error");
      vi.mocked(sessionsStore.closeSession).mockRejectedValueOnce(error);

      const listItem = wrapper.find(".v-list-item");
      await listItem.trigger("click");

      const dialog = wrapper.findComponent({ name: "MessageDialog" });
      await dialog.vm.$emit("confirm");
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to close session.");
      expect(handleError).toHaveBeenCalledWith(error);
    });
  });
});
