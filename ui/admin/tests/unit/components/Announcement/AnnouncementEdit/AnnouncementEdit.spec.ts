import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import useAnnouncementStore from "@admin/store/modules/announcement";
import { SnackbarPlugin } from "@/plugins/snackbar";
import AnnouncementEdit from "../../../../../src/components/Announcement/AnnouncementEdit.vue";
import routes from "../../../../../src/router";

type AnnouncementEditWrapper = VueWrapper<InstanceType<typeof AnnouncementEdit>>;

const announcement = {
  uuid: "eac7e18d-7127-41ca-b68b-8242dfdbaf4c",
  title: "Announcement 1",
  content: "## ShellHub new features \n - New feature 1 \n - New feature 2 \n - New feature 3",
  date: "2022-12-15T19:45:45.618Z",
};

const propAnnouncement = {
  uuid: "eac7e18d-7127-41ca-b68b-8242dfdbaf4c",
  title: "Announcement 1",
  date: "2022-12-15T19:45:45.618Z",
};

describe("Announcement Edit", () => {
  const vuetify = createVuetify();
  let wrapper: AnnouncementEditWrapper;

  beforeEach(() => {
    setActivePinia(createPinia());

    const store = useAnnouncementStore();
    store.announcement = announcement;
    vi.spyOn(store, "fetchAnnouncement").mockImplementation(async () => {
      store.announcement = announcement;
    });

    wrapper = mount(AnnouncementEdit, {
      global: {
        plugins: [vuetify, routes, SnackbarPlugin],
      },
      props: {
        announcementItem: propAnnouncement,
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
    expect(wrapper.vm.dialog).toBe(false);
    expect(wrapper.vm.announcement.uuid).toBe(propAnnouncement.uuid);
    expect(wrapper.vm.announcement.title).toBe(propAnnouncement.title);
    expect(wrapper.vm.announcement.date).toBe(propAnnouncement.date);
    expect(wrapper.vm.contentInHtml).toBe("");
    expect(wrapper.vm.contentError).toBe(false);
    expect(wrapper.vm.title).toBe(announcement.title);
  });
});
