import { createStore } from "vuex";
import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import AnnouncementDelete from "../../../../../src/components/Announcement/AnnouncementDelete.vue";
import { key } from "../../../../../src/store";
import routes from "../../../../../src/router";

const announcement = {
  uuid: "eac7e18d-7127-41ca-b68b-8242dfdbaf4c",
  title: "Announcement 1",
  content: "## ShellHub new features \n - New feature 1 \n - New feature 2 \n - New feature 3",
  date: "2022-12-15T19:45:45.618Z",
};

describe("Announcement Delete", () => {
  const store = createStore({
    state: {
      announcement,
    },
    getters: {
      "announcement/announcement": () => announcement,
    },
    actions: {
      "announcement/getAnnouncement": vi.fn(),
      "announcement/announcement": vi.fn(),
    },
  });

  const vuetify = createVuetify();

  let wrapper: VueWrapper<any>;

  beforeEach(() => {
    wrapper = mount(AnnouncementDelete, {
      global: {
        plugins: [[store, key], vuetify, routes],
      },
      props: {
        uuid: "eac7e18d-7127-41ca-b68b-8242dfdbaf4c",
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
    expect(wrapper.vm.dialog).toBe(false);
  });
});
