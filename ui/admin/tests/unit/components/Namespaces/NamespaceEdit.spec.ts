import { describe, expect, it, beforeEach, vi, afterEach } from "vitest";
import { DOMWrapper, VueWrapper, flushPromises } from "@vue/test-utils";
import { mountComponent, mockSnackbar } from "@tests/utils/mount";
import { createAxiosError } from "@tests/utils/axiosError";
import useNamespacesStore from "@admin/store/modules/namespaces";
import NamespaceEdit from "@admin/components/Namespace/NamespaceEdit.vue";
import { mockNamespace } from "../../mocks";

describe("NamespaceEdit", () => {
  let wrapper: VueWrapper<InstanceType<typeof NamespaceEdit>>;
  let namespacesStore: ReturnType<typeof useNamespacesStore>;

  const mountWrapper = () => {
    wrapper = mountComponent(NamespaceEdit, {
      props: {
        namespace: mockNamespace,
        modelValue: true,
      },
      attachTo: document.body,
    });

    namespacesStore = useNamespacesStore();
  };

  const getDialog = () => new DOMWrapper(document.body).find('[role="dialog"]');

  afterEach(() => {
    vi.clearAllMocks();
    wrapper?.unmount();
    document.body.innerHTML = "";
  });

  describe("rendering", () => {
    beforeEach(() => mountWrapper());

    it("shows the dialog when modelValue is true", async () => {
      await flushPromises();
      const dialog = getDialog();

      expect(dialog.exists()).toBe(true);
      expect(dialog.text()).toContain("Edit Namespace");
    });

    it("displays current namespace values in form", async () => {
      await flushPromises();
      const dialog = getDialog();

      const nameInput = dialog.find('[data-test="name-text"] input');
      expect((nameInput.element as HTMLInputElement).value).toBe(mockNamespace.name);

      const maxDevicesInput = dialog.find('[data-test="maxDevices-text"] input');
      expect((maxDevicesInput.element as HTMLInputElement).value).toBe(mockNamespace.max_devices.toString());
    });

    it("shows save and cancel buttons", async () => {
      await flushPromises();
      const dialog = getDialog();

      expect(dialog.find('[data-test="confirm-btn"]').exists()).toBe(true);
      expect(dialog.find('[data-test="cancel-btn"]').exists()).toBe(true);
    });
  });

  describe("form validation", () => {
    beforeEach(() => mountWrapper());

    it("shows error when name is empty", async () => {
      await flushPromises();
      const dialog = getDialog();

      const nameInput = dialog.find('[data-test="name-text"] input');
      await nameInput.setValue("");
      await flushPromises();

      expect(dialog.text()).toContain("this is a required field");
    });

    it("shows error when max devices is below minimum", async () => {
      await flushPromises();
      const dialog = getDialog();

      const maxDevicesInput = dialog.find('[data-test="maxDevices-text"] input');
      await maxDevicesInput.setValue("-2");
      await flushPromises();

      expect(dialog.text()).toContain("Maximum devices must be -1 (unlimited) or greater");
    });

    it("accepts -1 for unlimited devices", async () => {
      await flushPromises();
      const dialog = getDialog();

      const maxDevicesInput = dialog.find('[data-test="maxDevices-text"] input');
      await maxDevicesInput.setValue("-1");
      await flushPromises();

      expect(dialog.text()).not.toContain("Maximum devices must be -1");
    });

    it("disables save button when form has errors", async () => {
      await flushPromises();
      const dialog = getDialog();

      const nameInput = dialog.find('[data-test="name-text"] input');
      await nameInput.setValue("");
      await flushPromises();

      const saveBtn = dialog.find('[data-test="confirm-btn"]');
      expect(saveBtn.attributes("disabled")).toBeDefined();
    });
  });

  describe("updating namespace", () => {
    it("calls store action with updated values on submit", async () => {
      mountWrapper();
      await flushPromises();
      const dialog = getDialog();

      const nameInput = dialog.find('[data-test="name-text"] input');
      await nameInput.setValue("updated-namespace");
      await flushPromises();

      const maxDevicesInput = dialog.find('[data-test="maxDevices-text"] input');
      await maxDevicesInput.setValue("42");
      await flushPromises();

      const saveBtn = dialog.find('[data-test="confirm-btn"]');
      await saveBtn.trigger("click");
      await flushPromises();

      expect(namespacesStore.updateNamespace).toHaveBeenCalledWith(
        expect.objectContaining({
          name: "updated-namespace",
          max_devices: 42,
          settings: expect.objectContaining({
            session_record: mockNamespace.settings.session_record,
          }),
        }),
      );
    });

    it("updates session record setting", async () => {
      mountWrapper();
      await flushPromises();
      const dialog = getDialog();

      // Find and toggle session record switch
      const sessionRecordSwitch = dialog.find('input[type="checkbox"]');
      const currentValue = (sessionRecordSwitch.element as HTMLInputElement).checked;
      await sessionRecordSwitch.setValue(!currentValue);
      await flushPromises();

      const saveBtn = dialog.find('[data-test="confirm-btn"]');
      await saveBtn.trigger("click");
      await flushPromises();

      expect(namespacesStore.updateNamespace).toHaveBeenCalledWith(
        expect.objectContaining({
          settings: expect.objectContaining({
            session_record: !currentValue,
          }),
        }),
      );
    });

    it("shows success message and closes dialog after successful update", async () => {
      mountWrapper();
      await flushPromises();
      const dialog = getDialog();

      const saveBtn = dialog.find('[data-test="confirm-btn"]');
      await saveBtn.trigger("click");
      await flushPromises();

      expect(mockSnackbar.showSuccess).toHaveBeenCalledWith("Namespace updated successfully.");
      expect(wrapper.emitted("update:modelValue")).toBeTruthy();
      expect(wrapper.emitted("update:modelValue")?.[0]).toEqual([false]);
    });

    it("emits update event after successful update", async () => {
      mountWrapper();
      await flushPromises();
      const dialog = getDialog();

      const saveBtn = dialog.find('[data-test="confirm-btn"]');
      await saveBtn.trigger("click");
      await flushPromises();

      expect(wrapper.emitted("update")).toBeTruthy();
    });

    it("shows error message when update fails", async () => {
      mountWrapper();
      vi.mocked(namespacesStore.updateNamespace).mockRejectedValueOnce(
        createAxiosError(500, "Internal Server Error"),
      );
      await flushPromises();
      const dialog = getDialog();

      const saveBtn = dialog.find('[data-test="confirm-btn"]');
      await saveBtn.trigger("click");
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to update namespace.");
      expect(wrapper.emitted("update:modelValue")?.[0]).toBeUndefined();
    });
  });

  describe("closing dialog", () => {
    beforeEach(() => mountWrapper());

    it("closes dialog when cancel button is clicked", async () => {
      await flushPromises();
      const dialog = getDialog();

      const cancelBtn = dialog.find('[data-test="cancel-btn"]');
      await cancelBtn.trigger("click");
      await flushPromises();

      expect(wrapper.emitted("update:modelValue")).toBeTruthy();
      expect(wrapper.emitted("update:modelValue")?.[0]).toEqual([false]);
    });

    it("resets form fields when dialog is closed", async () => {
      await flushPromises();
      const dialog = getDialog();

      // Change a field
      const nameInput = dialog.find('[data-test="name-text"] input');
      await nameInput.setValue("changed-name");
      await flushPromises();

      // Close dialog
      const cancelBtn = dialog.find('[data-test="cancel-btn"]');
      await cancelBtn.trigger("click");
      await flushPromises();

      // Reopen dialog by setting modelValue to true again
      await wrapper.setProps({ modelValue: true });
      await flushPromises();

      // Check that field is reset to original value
      const dialogReopened = getDialog();
      const nameInputReopened = dialogReopened.find('[data-test="name-text"] input');
      expect((nameInputReopened.element as HTMLInputElement).value).toBe(mockNamespace.name);
    });
  });
});
