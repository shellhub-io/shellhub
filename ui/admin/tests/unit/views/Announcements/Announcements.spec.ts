import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import { useAnnouncementStore } from "@admin/store/modules/announcement";
import routes from "../../../../src/router";
import Announcements from "../../../../src/views/Announcements.vue";

type AnnouncementsWrapper = VueWrapper<InstanceType<typeof Announcements>>;

const announcements = [
  {
    uuid: "eac7e18d-7127-41ca-b68b-8242dfdbaf4c",
    title: "Announcement 1",
    content: "## ShellHub new features \n - New feature 1 \n - New feature 2 \n - New feature 3",
    date: "2022-12-15T19:45:45.618Z",
  },
];

const numberAnnouncements = 1;

describe("Announcement Details", () => {
  let wrapper: AnnouncementsWrapper;

  beforeEach(() => {
    const pinia = createPinia();
    setActivePinia(pinia);

    const announcementStore = useAnnouncementStore();
    announcementStore.announcements = announcements;
    announcementStore.numberAnnouncements = numberAnnouncements;

    const vuetify = createVuetify();

    wrapper = mount(Announcements, {
      global: {
        plugins: [pinia, vuetify, routes],
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Has the correct data", () => {
    expect(wrapper.find("[data-test='announcement-title']").text()).toBe("Announcements");
  });

  it("Renders the correct HTML", () => {
    expect(wrapper.find("[data-test='announcement-title']").exists()).toBeTruthy();
    expect(wrapper.find("[data-test='new-announcement-btn']").exists()).toBeTruthy();
    expect(wrapper.find("[data-test='announcement-list']").exists()).toBeTruthy();
  });
});
