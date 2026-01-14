import { describe, expect, it, beforeEach, vi, afterEach } from "vitest";
import { VueWrapper, flushPromises } from "@vue/test-utils";
import { mountComponent, mockSnackbar } from "@tests/utils/mount";
import { createCleanAdminRouter } from "@tests/utils/router";
import { createAxiosError } from "@tests/utils/axiosError";
import useAnnouncementStore from "@admin/store/modules/announcement";
import AnnouncementDetails from "@admin/views/AnnouncementDetails.vue";
import { mockAnnouncement } from "../mocks";
import { formatFullDateTime } from "@/utils/date";

vi.mock("@admin/store/api/announcement");

describe("AnnouncementDetails", () => {
  let wrapper: VueWrapper<InstanceType<typeof AnnouncementDetails>>;

  const mountWrapper = async (mockError?: Error) => {
    const router = createCleanAdminRouter();
    await router.push({ name: "announcementDetails", params: { uuid: mockAnnouncement.uuid } });
    await router.isReady();

    wrapper = mountComponent(AnnouncementDetails, {
      global: { plugins: [router] },
      piniaOptions: {
        initialState: { adminAnnouncement: mockError ? {} : { announcement: mockAnnouncement } },
        stubActions: !mockError,
      },
    });

    const announcementStore = useAnnouncementStore();
    if (mockError) vi.mocked(announcementStore.fetchAnnouncement).mockRejectedValueOnce(mockError);

    await flushPromises();
  };

  afterEach(() => {
    vi.clearAllMocks();
    wrapper?.unmount();
  });

  describe("when announcement loads successfully", () => {
    beforeEach(() => mountWrapper());

    it("displays announcement title in card header", () => {
      expect(wrapper.find(".text-h6").text()).toBe(mockAnnouncement.title);
    });

    it("displays uuid field with value", () => {
      const uuidField = wrapper.find('[data-test="announcement-uuid-field"]');
      expect(uuidField.text()).toContain("UUID:");
      expect(uuidField.text()).toContain(mockAnnouncement.uuid);
    });

    it("displays date field", () => {
      const dateField = wrapper.find('[data-test="announcement-date-field"]');
      expect(dateField.text()).toContain("Date:");
      expect(dateField.text()).toContain(formatFullDateTime(mockAnnouncement.date));
    });

    it("displays content field", () => {
      const contentField = wrapper.find('[data-test="announcement-content-field"]');
      expect(contentField.text()).toContain("Content:");
      expect(contentField.html()).toContain("<h2>ShellHub new features</h2>");
    });

    it("shows actions menu button", () => {
      const menuBtn = wrapper.find('[data-test="announcement-actions-menu-btn"]');
      expect(menuBtn.exists()).toBe(true);
    });
  });

  describe("when announcement fails to load", () => {
    it("shows error snackbar", async () => {
      await mountWrapper(createAxiosError(404, "Not Found"));
      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to get announcement details.");
    });
  });
});
