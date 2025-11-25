import { createVuetify } from "vuetify";
import { flushPromises, mount } from "@vue/test-utils";
import { describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import useAnnouncementStore from "@admin/store/modules/announcement";
import routes from "@admin/router";
import AnnouncementDetails from "@admin/views/AnnouncementDetails.vue";
import { SnackbarPlugin } from "@/plugins/snackbar";

const mockAnnouncement = {
  uuid: "eac7e18d-7127-41ca-b68b-8242dfdbaf4c",
  title: "Announcement 1",
  content: "## ShellHub new features \n - New feature 1 \n - New feature 2 \n - New feature 3",
  date: "2022-12-15T19:45:45.618Z",
};

const announcementContentInHtml = `<h2>ShellHub new features</h2>
<ul>
<li>New feature 1</li>
<li>New feature 2</li>
<li>New feature 3</li>
</ul>
`;

const mockRoute = { params: { uuid: mockAnnouncement.uuid } };

describe("Announcement Details", async () => {
  setActivePinia(createPinia());
  const announcementStore = useAnnouncementStore();

  announcementStore.fetchAnnouncement = vi.fn().mockResolvedValue(mockAnnouncement);
  announcementStore.announcement = mockAnnouncement;

  const wrapper = mount(AnnouncementDetails, {
    global: {
      plugins: [createVuetify(), routes, SnackbarPlugin],
      mocks: {
        $route: mockRoute,
        $router: { push: vi.fn() },
      },
    },
  });

  await flushPromises();

  it("Displays announcement title in card header", () => {
    expect(wrapper.find(".text-h6").text()).toBe(mockAnnouncement.title);
  });

  it("Displays UUID field", () => {
    const uuidField = wrapper.find('[data-test="announcement-uuid-field"]');
    expect(uuidField.text()).toContain("UUID:");
    expect(uuidField.text()).toContain(mockAnnouncement.uuid);
  });

  it("Displays date field with formatted date", () => {
    const dateField = wrapper.find('[data-test="announcement-date-field"]');
    expect(dateField.text()).toContain("Date:");
  });

  it("Displays content field", () => {
    const contentField = wrapper.find('[data-test="announcement-content-field"]');
    expect(contentField.text()).toContain("Content:");
  });

  it("Renders markdown content as HTML", () => {
    expect(wrapper.vm.contentToHtml.trim()).toBe(announcementContentInHtml.trim());
  });

  it("Shows actions menu button", () => {
    const menuBtn = wrapper.find('[data-test="announcement-actions-menu-btn"]');
    expect(menuBtn.exists()).toBe(true);
  });
});
