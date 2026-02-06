import { describe, expect, it, beforeEach, afterEach, vi } from "vitest";
import { VueWrapper, flushPromises } from "@vue/test-utils";
import { mockSnackbar, mountComponent } from "@tests/utils/mount";
import { createAxiosError } from "@tests/utils/axiosError";
import SettingNamespace from "@/components/Setting/SettingNamespace.vue";
import useNamespacesStore from "@/store/modules/namespaces";
import handleError from "@/utils/handleError";
import { mockNamespace } from "@tests/mocks";
import * as hasPermissionModule from "@/utils/permission";
import { createCleanRouter } from "@tests/utils/router";

describe("SettingNamespace", () => {
  let wrapper: VueWrapper<InstanceType<typeof SettingNamespace>>;
  let namespacesStore: ReturnType<typeof useNamespacesStore>;

  const mountWrapper = ({
    isOwner = true,
    hasPermission = true,
    namespace = mockNamespace,
  } = {}) => {
    vi.spyOn(hasPermissionModule, "default").mockReturnValue(hasPermission);

    const ownerId = isOwner ? "test" : "user-456";
    localStorage.setItem("id", ownerId);
    localStorage.setItem("tenant", namespace.tenant_id);

    wrapper = mountComponent(SettingNamespace, {
      global: { plugins: [createCleanRouter()] },
      piniaOptions: {
        initialState: {
          namespaces: {
            namespaces: [namespace],
            currentNamespace: namespace,
          },
          auth: { tenantId: namespace.tenant_id },
        },
      },
    });

    namespacesStore = useNamespacesStore();
  };

  beforeEach(() => {
    mountWrapper();
  });

  afterEach(() => {
    wrapper?.unmount();
    vi.clearAllMocks();
  });

  describe("Page header", () => {
    it("Renders page header with correct props", () => {
      const header = wrapper.findComponent({ name: "PageHeader" });
      expect(header.exists()).toBe(true);
      expect(header.props("icon")).toBe("mdi-cloud-braces");
      expect(header.props("title")).toBe("Namespace");
      expect(header.props("overline")).toBe("Settings");
    });

    it("Shows edit button when not in edit mode", () => {
      const editBtn = wrapper.find('[data-test="edit-namespace-btn"]');
      expect(editBtn.exists()).toBe(true);
      expect(editBtn.text()).toContain("Edit Namespace");
    });

    it("Shows cancel and save buttons when in edit mode", async () => {
      await wrapper.find('[data-test="edit-namespace-btn"]').trigger("click");

      const cancelBtn = wrapper.find('[data-test="cancel-edit-btn"]');
      const saveBtn = wrapper.find('[data-test="save-changes-btn"]');

      expect(cancelBtn.exists()).toBe(true);
      expect(saveBtn.exists()).toBe(true);
      expect(wrapper.find('[data-test="edit-namespace-btn"]').exists()).toBe(false);
    });
  });

  describe("Name field", () => {
    it("Renders name input with current namespace name", () => {
      const nameInput = wrapper.find('[data-test="name-input"] input');
      expect(nameInput.exists()).toBe(true);
      expect((nameInput.element as HTMLInputElement).value).toBe(mockNamespace.name);
    });

    it("Name input is readonly when not in edit mode", () => {
      const nameInput = wrapper.findComponent({ name: "v-text-field", props: { modelValue: mockNamespace.name } });
      expect(nameInput.props("readonly")).toBe(true);
      expect(nameInput.props("disabled")).toBe(true);
    });

    it("Name input is editable when in edit mode", async () => {
      const editBtn = wrapper.find('[data-test="edit-namespace-btn"]');
      await editBtn.trigger("click");

      const vTextField = wrapper.findComponent({ name: "v-text-field" });
      expect(vTextField.props("readonly")).toBe(false);
      expect(vTextField.props("disabled")).toBe(false);
    });

    it("Shows error when name is too short", async () => {
      const editBtn = wrapper.find('[data-test="edit-namespace-btn"]');
      await editBtn.trigger("click");

      const nameInput = wrapper.find('[data-test="name-input"] input');
      await nameInput.setValue("ab");
      await flushPromises();

      const vTextField = wrapper.findComponent({ name: "v-text-field" });
      expect(vTextField.props("errorMessages")).toBeTruthy();
    });

    it("Shows error when name contains dots", async () => {
      const editBtn = wrapper.find('[data-test="edit-namespace-btn"]');
      await editBtn.trigger("click");

      const nameInput = wrapper.find('[data-test="name-input"] input');
      await nameInput.setValue("test.namespace");
      await flushPromises();

      const vTextField = wrapper.findComponent({ name: "v-text-field" });
      expect(vTextField.props("errorMessages")).toBeTruthy();
    });

    it("Disables save button when name has error", async () => {
      const editBtn = wrapper.find('[data-test="edit-namespace-btn"]');
      await editBtn.trigger("click");

      const nameInput = wrapper.find('[data-test="name-input"] input');
      await nameInput.setValue("ab");
      await flushPromises();

      const saveBtn = wrapper.find('[data-test="save-changes-btn"]');
      expect(saveBtn.attributes("disabled")).toBeDefined();
    });
  });

  describe("Namespace type", () => {
    it("Renders type chip with namespace type", () => {
      const typeChip = wrapper.find('[data-test="type-chip"]');
      expect(typeChip.exists()).toBe(true);
      expect(typeChip.text()).toContain("personal");
    });

    it("Shows personal icon for personal type", () => {
      const typeChip = wrapper.find('[data-test="type-chip"]');
      const icon = typeChip.find(".v-icon");
      expect(icon.classes().join(" ")).toContain("mdi-account");
    });

    it("Shows team icon for team type", async () => {
      wrapper.unmount();
      const teamNamespace = { ...mockNamespace, type: "team" as const };
      mountWrapper({ namespace: teamNamespace });
      await flushPromises();

      const typeChip = wrapper.find('[data-test="type-chip"]');
      const icon = typeChip.find(".v-icon");
      expect(icon.classes().join(" ")).toContain("mdi-account-group");
    });
  });

  describe("Tenant ID", () => {
    it("Renders tenant ID chip", () => {
      const tenantCopy = wrapper.find('[data-test="tenant-copy-btn"]');
      expect(tenantCopy.exists()).toBe(true);
      expect(tenantCopy.text()).toContain(mockNamespace.tenant_id);
    });

    it("Shows copy icon", () => {
      const tenantCopy = wrapper.find('[data-test="tenant-copy-btn"]');
      const icon = tenantCopy.find(".v-icon");
      expect(icon.exists()).toBe(true);
      expect(icon.classes().join(" ")).toContain("mdi-content-copy");
    });
  });

  describe("Connection announcement", () => {
    it("Renders announcement row", () => {
      const announcementRow = wrapper.find('[data-test="announcement-item"]');
      expect(announcementRow.exists()).toBe(true);
    });

    it("Renders edit announcement button", () => {
      const editBtn = wrapper.find('[data-test="edit-announcement-btn"]');
      expect(editBtn.exists()).toBe(true);
      expect(editBtn.text()).toContain("Edit Announcement");
    });

    it("Opens announcement dialog when button is clicked", async () => {
      const editBtn = wrapper.find('[data-test="edit-announcement-btn"]');
      await editBtn.trigger("click");

      const dialog = wrapper.findComponent({ name: "ConnectionAnnouncementEdit" });
      expect(dialog.props("modelValue")).toBe(true);
    });
  });

  describe("Session recording", () => {
    it("Renders session recording component", () => {
      const sessionRecording = wrapper.find('[data-test="session-recording-setting-component"]');
      expect(sessionRecording.exists()).toBe(true);
    });

    it("Passes tenant ID to session recording component", () => {
      const sessionRecording = wrapper.findComponent({ name: "SettingSessionRecording" });
      expect(sessionRecording.props("tenantId")).toBe(mockNamespace.tenant_id);
    });
  });

  describe("Delete/Leave namespace", () => {
    it("Shows delete button when user is owner", () => {
      const deleteBtn = wrapper.find('[data-test="delete-namespace-btn"]');
      expect(deleteBtn.exists()).toBe(true);
      expect(deleteBtn.text()).toContain("Delete");
    });

    it("Shows leave button when user is not owner", async () => {
      wrapper.unmount();
      mountWrapper({ isOwner: false });
      await flushPromises();

      const leaveBtn = wrapper.find('[data-test="leave-namespace-btn"]');
      expect(leaveBtn.exists()).toBe(true);
      expect(leaveBtn.text()).toContain("Leave");
    });

    it("Opens delete dialog when delete button is clicked", async () => {
      const deleteBtn = wrapper.find('[data-test="delete-namespace-btn"]');
      await deleteBtn.trigger("click");

      const dialog = wrapper.findComponent({ name: "NamespaceDelete" });
      expect(dialog.props("modelValue")).toBe(true);
    });

    it("Opens leave dialog when leave button is clicked", async () => {
      wrapper.unmount();
      mountWrapper({ isOwner: false });
      await flushPromises();

      const leaveBtn = wrapper.find('[data-test="leave-namespace-btn"]');
      await leaveBtn.trigger("click");

      const dialog = wrapper.findComponent({ name: "NamespaceLeave" });
      expect(dialog.props("modelValue")).toBe(true);
    });
  });

  describe("Edit and cancel", () => {
    it("Enters edit mode when edit button is clicked", async () => {
      const editBtn = wrapper.find('[data-test="edit-namespace-btn"]');
      await editBtn.trigger("click");

      const nameInput = wrapper.findComponent({ name: "v-text-field" });
      expect(nameInput.props("readonly")).toBe(false);
    });

    it("Exits edit mode and resets name when cancel is clicked", async () => {
      const editBtn = wrapper.find('[data-test="edit-namespace-btn"]');
      await editBtn.trigger("click");

      const nameInput = wrapper.find('[data-test="name-input"] input');
      await nameInput.setValue("new-name");

      const cancelBtn = wrapper.find('[data-test="cancel-edit-btn"]');
      await cancelBtn.trigger("click");

      const resetInput = wrapper.find('[data-test="name-input"] input');
      expect((resetInput.element as HTMLInputElement).value).toBe(mockNamespace.name);
    });
  });

  describe("Update name", () => {
    it("Calls editNamespace with correct data", async () => {
      const editBtn = wrapper.find('[data-test="edit-namespace-btn"]');
      await editBtn.trigger("click");

      const nameInput = wrapper.find('[data-test="name-input"] input');
      await nameInput.setValue("new-namespace");

      const saveBtn = wrapper.find('[data-test="save-changes-btn"]');
      await saveBtn.trigger("click");
      await flushPromises();

      expect(namespacesStore.editNamespace).toHaveBeenCalledWith({
        tenant_id: mockNamespace.tenant_id,
        name: "new-namespace",
      });
    });

    it("Fetches namespace list after update", async () => {
      const editBtn = wrapper.find('[data-test="edit-namespace-btn"]');
      await editBtn.trigger("click");

      const nameInput = wrapper.find('[data-test="name-input"] input');
      await nameInput.setValue("new-namespace");

      const saveBtn = wrapper.find('[data-test="save-changes-btn"]');
      await saveBtn.trigger("click");
      await flushPromises();

      expect(namespacesStore.fetchNamespaceList).toHaveBeenCalled();
    });

    it("Fetches updated namespace after update", async () => {
      const editBtn = wrapper.find('[data-test="edit-namespace-btn"]');
      await editBtn.trigger("click");

      const nameInput = wrapper.find('[data-test="name-input"] input');
      await nameInput.setValue("new-namespace");

      const saveBtn = wrapper.find('[data-test="save-changes-btn"]');
      await saveBtn.trigger("click");
      await flushPromises();

      expect(namespacesStore.fetchNamespace).toHaveBeenCalledWith(mockNamespace.tenant_id);
    });

    it("Shows success message after update", async () => {
      const editBtn = wrapper.find('[data-test="edit-namespace-btn"]');
      await editBtn.trigger("click");

      const nameInput = wrapper.find('[data-test="name-input"] input');
      await nameInput.setValue("new-namespace");

      const saveBtn = wrapper.find('[data-test="save-changes-btn"]');
      await saveBtn.trigger("click");
      await flushPromises();

      expect(mockSnackbar.showSuccess).toHaveBeenCalledWith("Namespace name updated successfully.");
    });

    it("Exits edit mode after successful update", async () => {
      const editBtn = wrapper.find('[data-test="edit-namespace-btn"]');
      await editBtn.trigger("click");

      const nameInput = wrapper.find('[data-test="name-input"] input');
      await nameInput.setValue("new-namespace");

      const saveBtn = wrapper.find('[data-test="save-changes-btn"]');
      await saveBtn.trigger("click");
      await flushPromises();

      const editBtnAfter = wrapper.find('[data-test="edit-namespace-btn"]');
      expect(editBtnAfter.exists()).toBe(true);
    });
  });

  describe("Update errors", () => {
    beforeEach(() => wrapper?.unmount());

    it("Shows error message for 400 error", async () => {
      const error = createAxiosError(400, "Bad Request");

      mountWrapper();
      vi.mocked(namespacesStore.editNamespace).mockRejectedValueOnce(error);
      await flushPromises();

      const editBtn = wrapper.find('[data-test="edit-namespace-btn"]');
      await editBtn.trigger("click");

      const nameInput = wrapper.find('[data-test="name-input"] input');
      await nameInput.setValue("new-namespace");
      await flushPromises();

      const saveBtn = wrapper.find('[data-test="save-changes-btn"]');
      await saveBtn.trigger("click");
      await flushPromises();

      const vTextField = wrapper.findComponent({ name: "v-text-field" });
      expect(vTextField.props("errorMessages")).toBe("This name is not valid");
      expect(vTextField.props("hideDetails")).toBe(false);
    });

    it("Shows error message for 409 error", async () => {
      const error = createAxiosError(409, "Conflict");

      mountWrapper();
      vi.mocked(namespacesStore.editNamespace).mockRejectedValueOnce(error);
      await flushPromises();

      const editBtn = wrapper.find('[data-test="edit-namespace-btn"]');
      await editBtn.trigger("click");

      const nameInput = wrapper.find('[data-test="name-input"] input');
      await nameInput.setValue("existing-namespace");
      await flushPromises();

      const saveBtn = wrapper.find('[data-test="save-changes-btn"]');
      await saveBtn.trigger("click");
      await flushPromises();

      const vTextField = wrapper.findComponent({ name: "v-text-field" });
      expect(vTextField.props("errorMessages")).toBe("Name used already");
      expect(vTextField.props("hideDetails")).toBe(false);
    });

    it("Shows generic error for other errors", async () => {
      const error = createAxiosError(500, "Internal Server Error");

      mountWrapper();

      vi.mocked(namespacesStore.editNamespace).mockRejectedValueOnce(error);
      await flushPromises();

      const editBtn = wrapper.find('[data-test="edit-namespace-btn"]');
      await editBtn.trigger("click");

      const nameInput = wrapper.find('[data-test="name-input"] input');
      await nameInput.setValue("new-namespace");

      const saveBtn = wrapper.find('[data-test="save-changes-btn"]');
      await saveBtn.trigger("click");
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to update name.");
      expect(handleError).toHaveBeenCalledWith(error);
    });
  });

  describe("Namespace fetch on mount", () => {
    it("Fetches namespace on mount", async () => {
      await flushPromises();
      expect(namespacesStore.fetchNamespace).toHaveBeenCalledWith(mockNamespace.tenant_id);
    });

    it("Shows error when fetch fails with 403", async () => {
      const error = createAxiosError(403, "Forbidden");

      wrapper.unmount();
      mountWrapper();
      vi.mocked(namespacesStore.fetchNamespace).mockRejectedValueOnce(error);
      await flushPromises();

      await wrapper.findComponent({ name: "ConnectionAnnouncementEdit" }).vm.$emit("update");
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("You are not authorized to access this resource.");
    });

    it("Shows generic error when fetch fails", async () => {
      const error = createAxiosError(500, "Internal Server Error");

      wrapper.unmount();
      mountWrapper();
      vi.mocked(namespacesStore.fetchNamespace).mockRejectedValueOnce(error);
      await flushPromises();

      await wrapper.findComponent({ name: "ConnectionAnnouncementEdit" }).vm.$emit("update");
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to load namespace.");
      expect(handleError).toHaveBeenCalledWith(error);
    });
  });
});
