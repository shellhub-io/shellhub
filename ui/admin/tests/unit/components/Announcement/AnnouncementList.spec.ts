import { describe, expect, it, beforeEach, vi, afterEach } from "vitest";
import { VueWrapper, flushPromises } from "@vue/test-utils";
import { mountComponent, mockSnackbar } from "@tests/utils/mount";
import { createCleanAdminRouter } from "@tests/utils/router";
import { createAxiosError } from "@tests/utils/axiosError";
import useAnnouncementStore from "@admin/store/modules/announcement";
import AnnouncementList from "@admin/components/Announcement/AnnouncementList.vue";
import { mockAnnouncements } from "../../mocks";
import { Router } from "vue-router";

describe("AnnouncementList", () => {
  let wrapper: VueWrapper<InstanceType<typeof AnnouncementList>>;
  let router: Router;
  let announcementsStore: ReturnType<typeof useAnnouncementStore>;

  const mountWrapper = (mockAnnouncementCount?: number) => {
    router = createCleanAdminRouter();

    wrapper = mountComponent(AnnouncementList, {
      global: { plugins: [router] },
      piniaOptions: {
        initialState: {
          adminAnnouncement: {
            announcements: mockAnnouncements,
            announcementCount: mockAnnouncementCount ?? mockAnnouncements.length,
          },
        },
      },
    });

    announcementsStore = useAnnouncementStore();
  };

  afterEach(() => {
    vi.clearAllMocks();
    wrapper?.unmount();
  });

  describe("rendering", () => {
    beforeEach(() => mountWrapper());

    it("renders the data table", () => {
      expect(wrapper.find('[data-test="announcement-list"]').exists()).toBe(true);
    });

    it("displays announcement UUIDs", () => {
      const uuids = wrapper.findAll('[data-test="announcement-uuid"]');
      expect(uuids).toHaveLength(mockAnnouncements.length);
      expect(uuids[0].text()).toContain(mockAnnouncements[0].uuid);
    });

    it("displays announcement titles", () => {
      const titles = wrapper.findAll('[data-test="announcement-title"]');
      expect(titles).toHaveLength(mockAnnouncements.length);
      expect(titles[0].text()).toBe(mockAnnouncements[0].title);
      expect(titles[1].text()).toBe(mockAnnouncements[1].title);
    });

    it("displays action buttons for each announcement", () => {
      const actionCells = wrapper.findAll('[data-test="announcement-actions"]');
      expect(actionCells).toHaveLength(mockAnnouncements.length);
    });

    it("displays edit buttons", () => {
      const editButtons = wrapper.findAll('[data-test="edit-button"]');
      expect(editButtons).toHaveLength(mockAnnouncements.length);
    });

    it("displays delete buttons", () => {
      const deleteButtons = wrapper.findAll('[data-test="delete-button"]');
      expect(deleteButtons).toHaveLength(mockAnnouncements.length);
    });
  });

  describe("fetching announcements", () => {
    it("fetches announcements on mount", () => {
      mountWrapper();

      expect(announcementsStore.fetchAnnouncementList).toHaveBeenCalledWith(
        expect.objectContaining({
          perPage: 10,
          page: 1,
          orderBy: "desc",
        }),
      );
    });

    it("refetches announcements when page changes", async () => {
      mountWrapper(11); // Mock total count to 11 to enable pagination

      // Click next page button
      const nextPageBtn = wrapper.find('[data-test="pager-next"]');
      await nextPageBtn.trigger("click");
      await flushPromises();

      expect(announcementsStore.fetchAnnouncementList).toHaveBeenCalledWith(
        expect.objectContaining({
          page: 2,
        }),
      );
    });

    it("refetches announcements when items per page changes", async () => {
      mountWrapper(20);

      // Change items per page via combobox
      const ippCombo = wrapper.find('[data-test="ipp-combo"] input');
      await ippCombo.setValue(20);
      await flushPromises();

      expect(announcementsStore.fetchAnnouncementList).toHaveBeenCalledWith(
        expect.objectContaining({
          perPage: 20,
        }),
      );
    });
  });

  describe("navigating to announcement details", () => {
    it("navigates when clicking info icon", async () => {
      mountWrapper();

      const pushSpy = vi.spyOn(router, "push");
      const infoIcon = wrapper.findAll('[data-test="info-button"]')[0];

      await infoIcon.trigger("click");

      expect(pushSpy).toHaveBeenCalledWith({
        name: "announcementDetails",
        params: { uuid: mockAnnouncements[0].uuid },
      });
    });
  });

  describe("error handling", () => {
    it("shows error snackbar when fetching announcements fails", async () => {
      mountWrapper();
      vi.mocked(announcementsStore.fetchAnnouncementList).mockRejectedValueOnce(
        createAxiosError(500, "Internal Server Error"),
      );
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to fetch announcements.");
    });
  });
});
