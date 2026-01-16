import { describe, expect, it, beforeEach, vi, afterEach } from "vitest";
import { DOMWrapper, VueWrapper, flushPromises } from "@vue/test-utils";
import { mountComponent, mockSnackbar } from "@tests/utils/mount";
import useAnnouncementStore from "@admin/store/modules/announcement";
import AnnouncementEdit from "@admin/components/Announcement/AnnouncementEdit.vue";
import { mockAnnouncement, mockAnnouncementShort } from "../../mocks";
import { createAxiosError } from "@tests/utils/axiosError";

vi.mock("@tinymce/tinymce-vue", () => ({
  default: {
    name: "Editor",
    // eslint-disable-next-line vue/max-len
    template: '<div class="tinymce-mock"><textarea :value="modelValue" @input="$emit(\'update:modelValue\', $event.target.value)" data-test="editor-mock"></textarea></div>',
    props: ["modelValue", "init", "apiKey"],
  },
}));

const triggerButtonTemplate = `
  <template #default="{ openDialog }">
    <button 
      data-test="trigger-button" 
      @click="openDialog"
    >
      Edit
    </button>
  </template>
`;

describe("AnnouncementEdit", () => {
  let wrapper: VueWrapper<InstanceType<typeof AnnouncementEdit>>;
  let announcementsStore: ReturnType<typeof useAnnouncementStore>;

  const mountWrapper = () => {
    wrapper = mountComponent(AnnouncementEdit, {
      props: {
        announcementItem: mockAnnouncementShort,
        showTooltip: false,
      },
      slots: { default: triggerButtonTemplate },
      attachTo: document.body,
      piniaOptions: {
        initialState: {
          adminAnnouncement: { announcement: mockAnnouncement },
        },
      },
    });

    announcementsStore = useAnnouncementStore();
  };

  const openDialog = async () => {
    await wrapper.find('[data-test="trigger-button"]').trigger("click");
    return new DOMWrapper(document.body).find('[role="dialog"]');
  };

  afterEach(() => {
    vi.clearAllMocks();
    wrapper?.unmount();
    document.body.innerHTML = "";
  });

  describe("rendering", () => {
    beforeEach(() => mountWrapper());

    it("renders the slot content", () => {
      const trigger = wrapper.find('[data-test="trigger-button"]');
      expect(trigger.exists()).toBe(true);
      expect(trigger.text()).toBe("Edit");
    });

    it("does not show the dialog initially", () => {
      expect(new DOMWrapper(document.body).find('[role="dialog"]').exists()).toBe(false);
    });
  });

  describe("opening dialog", () => {
    it("shows dialog and fetches announcement when clicking trigger", async () => {
      mountWrapper();

      const dialog = await openDialog();

      expect(announcementsStore.fetchAnnouncement).toHaveBeenCalledWith(mockAnnouncementShort.uuid);
      expect(dialog.exists()).toBe(true);
      expect(dialog.text()).toContain("Edit Announcement");
    });

    it("displays current announcement title in form", async () => {
      mountWrapper();

      const dialog = await openDialog();
      const titleInput = dialog.find('input[type="text"]');
      expect((titleInput.element as HTMLInputElement).value).toBe(mockAnnouncement.title);
    });

    it("renders content editor", async () => {
      mountWrapper();

      const dialog = await openDialog();
      expect(dialog.find('[data-test="editor-mock"]').exists()).toBe(true);
    });
  });

  describe("form validation", () => {
    it("shows error when submitting with empty title", async () => {
      mountWrapper();

      const dialog = await openDialog();
      const titleInput = dialog.find('input[type="text"]');
      await titleInput.setValue("");
      await flushPromises();

      const submitBtn = dialog.find('[data-test="confirm-btn"]');
      await submitBtn.trigger("click");
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Please fill in all required fields.");
    });

    it("shows error when submitting with empty content", async () => {
      mountWrapper();

      const dialog = await openDialog();
      // Clear content via mocked editor
      const editor = dialog.find('[data-test="editor-mock"]');
      await editor.setValue("");
      await flushPromises();

      const submitBtn = dialog.find('[data-test="confirm-btn"]');
      await submitBtn.trigger("click");
      await flushPromises();

      expect(dialog.find('[data-test="announcement-error"]').exists()).toBe(true);
      expect(dialog.text()).toContain("The announcement cannot be empty!");
      expect(mockSnackbar.showError).toHaveBeenCalledWith("Please fill in all required fields.");
    });
  });

  describe("updating announcement", () => {
    it("calls store action and shows success message on submit", async () => {
      mountWrapper();

      const dialog = await openDialog();
      const titleInput = dialog.find('input[type="text"]');
      await titleInput.setValue("Updated Title");

      // Set content via mocked editor
      const editor = dialog.find('[data-test="editor-mock"]');
      await editor.setValue("<h2>Updated Content</h2>");
      await flushPromises();

      const submitBtn = dialog.find('[data-test="confirm-btn"]');
      await submitBtn.trigger("click");
      await flushPromises();

      expect(announcementsStore.updateAnnouncement).toHaveBeenCalledWith(
        mockAnnouncement.uuid,
        expect.objectContaining({
          title: "Updated Title",
          content: expect.stringContaining("Updated Content"),
        }),
      );
      expect(mockSnackbar.showSuccess).toHaveBeenCalledWith("Announcement updated successfully.");
    });

    it("shows error message when update fails", async () => {
      mountWrapper();
      vi.mocked(announcementsStore.updateAnnouncement).mockRejectedValueOnce(
        createAxiosError(500, "Internal Server Error"),
      );

      const dialog = await openDialog();
      const editor = dialog.find('[data-test="editor-mock"]');
      await editor.setValue("<h2>Updated Content</h2>");
      await flushPromises();

      const submitBtn = dialog.find('[data-test="confirm-btn"]');
      await submitBtn.trigger("click");
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to update announcement.");
    });

    it("emits update event after successful update", async () => {
      mountWrapper();

      const dialog = await openDialog();
      const editor = dialog.find('[data-test="editor-mock"]');
      await editor.setValue("<h2>Updated Content</h2>");
      await flushPromises();

      const submitBtn = dialog.find('[data-test="confirm-btn"]');
      await submitBtn.trigger("click");
      await flushPromises();

      expect(wrapper.emitted("update")).toBeTruthy();
    });
  });
});
