import { createStore } from "vuex";
import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { key } from "../../../../src/store";
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
  const store = createStore({
    state: {
      announcements,
    },
    getters: {
      "announcement/announcements": () => announcements,
      "announcement/numberAnnouncements": () => numberAnnouncements,
    },
    actions: {
      "announcement/getAnnouncements": vi.fn(),
      "announcement/announcements": vi.fn(),
    },
  });

  const vuetify = createVuetify();

  let wrapper: AnnouncementsWrapper;

  beforeEach(() => {
    wrapper = mount(Announcements, {
      global: {
        plugins: [[store, key], vuetify, routes],
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
