import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import useAnnouncementStore from "@admin/store/modules/announcement";
import AnnouncementEdit from "@admin/components/Announcement/AnnouncementEdit.vue";
import routes from "@admin/router";
import { SnackbarPlugin } from "@/plugins/snackbar";

type AnnouncementEditWrapper = VueWrapper<InstanceType<typeof AnnouncementEdit>>;

const shortAnnouncement = {
  uuid: "eac7e18d-7127-41ca-b68b-8242dfdbaf4c",
  title: "Announcement 1",
  date: "2022-12-15T19:45:45.618Z",
};

const announcement = {
  ...shortAnnouncement,
  content: "## ShellHub new features \n - New feature 1 \n - New feature 2 \n - New feature 3",
};

describe("Announcement Edit", () => {
  const vuetify = createVuetify();
  let wrapper: AnnouncementEditWrapper;

  beforeEach(() => {
    setActivePinia(createPinia());
    const announcementStore = useAnnouncementStore();

    announcementStore.announcement = announcement;
    vi.spyOn(announcementStore, "fetchAnnouncement").mockImplementation(async () => {
      announcementStore.announcement = announcement;
    });

    wrapper = mount(AnnouncementEdit, {
      global: {
        plugins: [vuetify, routes, SnackbarPlugin],
      },
      props: {
        announcementItem: shortAnnouncement,
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.exists()).toBe(true);
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders the correct data", () => {
    expect(wrapper.vm.showDialog).toBe(false);
    expect(wrapper.vm.announcement.uuid).toBe(shortAnnouncement.uuid);
    expect(wrapper.vm.announcement.title).toBe(shortAnnouncement.title);
    expect(wrapper.vm.announcement.date).toBe(shortAnnouncement.date);
    expect(wrapper.vm.contentInHtml).toBe("");
    expect(wrapper.vm.contentError).toBe(false);
    expect(wrapper.vm.title).toBe(announcement.title);
  });

  it("Opens the dialog on edit button click", async () => {
    expect(wrapper.vm.showDialog).toBe(false);

    const editButton = wrapper.find('[data-test="edit-button"]');
    await editButton.trigger("click");

    expect(wrapper.vm.showDialog).toBe(true);
  });
});
