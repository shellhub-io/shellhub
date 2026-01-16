import { describe, expect, it, beforeEach, vi, afterEach } from "vitest";
import { DOMWrapper, VueWrapper, flushPromises } from "@vue/test-utils";
import { mountComponent, mockSnackbar } from "@tests/utils/mount";
import { createCleanAdminRouter } from "@tests/utils/router";
import { createAxiosError } from "@tests/utils/axiosError";
import useAnnouncementStore from "@admin/store/modules/announcement";
import AnnouncementDelete from "@admin/components/Announcement/AnnouncementDelete.vue";
import { Router } from "vue-router";

const triggerButtonTemplate = `
  <template #default="{ openDialog }">
    <button 
      data-test="trigger-button" 
      @click="openDialog"
    >
      Delete
    </button>
  </template>
`;

describe("AnnouncementDelete", () => {
  let wrapper: VueWrapper<InstanceType<typeof AnnouncementDelete>>;
  let announcementsStore: ReturnType<typeof useAnnouncementStore>;
  let router: Router;
  const mockUuid = "announcement-123";

  const mountWrapper = (props: { redirect?: boolean } = {}) => {
    router = createCleanAdminRouter();

    wrapper = mountComponent(AnnouncementDelete, {
      global: { plugins: [router] },
      props: {
        uuid: mockUuid,
        ...props,
      },
      slots: {
        default: triggerButtonTemplate,
      },
      attachTo: document.body,
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
      expect(trigger.text()).toBe("Delete");
    });

    it("does not show the dialog initially", () => {
      expect(new DOMWrapper(document.body).find('[role="dialog"]').exists()).toBe(false);
    });
  });

  describe("opening dialog", () => {
    beforeEach(() => mountWrapper());

    it("shows dialog when clicking the trigger", async () => {
      const dialog = await openDialog();

      expect(dialog.exists()).toBe(true);
      expect(dialog.text()).toContain("Are you sure?");
      expect(dialog.text()).toContain("You are about to remove this announcement");
    });
  });

  describe("deleting announcement", () => {
    it("calls store action and shows success message on confirm", async () => {
      mountWrapper();

      const dialog = await openDialog();
      const confirmBtn = dialog.find('[data-test="confirm-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(announcementsStore.deleteAnnouncement).toHaveBeenCalledWith(mockUuid);
      expect(mockSnackbar.showSuccess).toHaveBeenCalledWith("Announcement deleted successfully.");
    });

    it("shows error message when delete fails", async () => {
      mountWrapper();
      vi.mocked(announcementsStore.deleteAnnouncement).mockRejectedValueOnce(
        createAxiosError(500, "Internal Server Error"),
      );

      const dialog = await openDialog();
      const confirmBtn = dialog.find('[data-test="confirm-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to delete announcement.");
    });

    it("redirects to announcements page when redirect prop is true", async () => {
      mountWrapper({ redirect: true });
      const pushSpy = vi.spyOn(router, "push");

      const dialog = await openDialog();
      const confirmBtn = dialog.find('[data-test="confirm-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(pushSpy).toHaveBeenCalledWith({ name: "announcements" });
    });

    it("emits update event when redirect prop is false", async () => {
      mountWrapper({ redirect: false });

      const dialog = await openDialog();
      const confirmBtn = dialog.find('[data-test="confirm-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(wrapper.emitted("update")).toBeTruthy();
    });
  });
});
