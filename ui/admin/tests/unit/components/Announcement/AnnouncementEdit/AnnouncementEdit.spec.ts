import { createStore } from "vuex";
import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import AnnouncementEdit from "../../../../../src/components/Announcement/AnnouncementEdit.vue";
import { key } from "../../../../../src/store";
import routes from "../../../../../src/router";

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
  const store = createStore({
    state: {
      announcement,
    },
    getters: {
      "announcement/announcement": () => announcement,
    },
    actions: {
      "announcement/updateAnnouncement": vi.fn(),
      "announcement/getAnnouncement": vi.fn(),
      "announcement/announcement": vi.fn(),
    },
  });

  const vuetify = createVuetify();

  let wrapper: VueWrapper<any>;

  beforeEach(() => {
    wrapper = mount(AnnouncementEdit, {
      global: {
        plugins: [[store, key], vuetify, routes],
      },
      props: {
        announcement: propAnnouncement,
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders the correct data", () => {
    expect(wrapper.vm.dialog).toBe(false);
    expect(wrapper.vm.announcement).toStrictEqual(propAnnouncement);
    expect(wrapper.vm.contentInHtml).toBe("");
    expect(wrapper.vm.contentError).toBe(false);
    expect(wrapper.vm.title).toBe(announcement.title);
  });
});
