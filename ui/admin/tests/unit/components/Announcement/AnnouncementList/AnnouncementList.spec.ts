import { createStore } from "vuex";
import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import AnnouncementList from "../../../../../src/components/Announcement/AnnouncementList.vue";
import { key } from "../../../../../src/store";
import routes from "../../../../../src/router";

type AnnouncementListWrapper = VueWrapper<InstanceType<typeof AnnouncementList>>;

const announcements = [
  {
    uuid: "eac7e18d-7127-41ca-b68b-8242dfdbaf4c",
    title: "Announcement 1",
    content: "## ShellHub new features \n - New feature 1 \n - New feature 2 \n - New feature 3",
    date: "2022-12-15T19:45:45.618Z",
  },
  {
    uuid: "eac7e18d-7127-41ca-b68b-8242dfdbaf5b",
    title: "Announcement 2",
    content: "## ShellHub new features \n - New feature 1 \n - New feature 2 \n - New feature 3",
    date: "2022-12-15T19:45:45.618Z",
  },
];

const numberAnnouncements = 2;

describe("Announcement List", () => {
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
      "snackbar/showSnackbarErrorAction": vi.fn(),
    },
  });

  const vuetify = createVuetify();

  let wrapper: AnnouncementListWrapper;

  beforeEach(() => {
    wrapper = mount(AnnouncementList, {
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

  it("Renders the correct data", () => {
    expect(wrapper.vm.itemsPerPage).toBe(10);
    expect(wrapper.vm.page).toBe(1);
    expect(wrapper.vm.loading).toBe(false);
  });

  it("Renders the correct computed", () => {
    expect(wrapper.vm.numberAnnouncements).toBe(numberAnnouncements);
    expect(wrapper.vm.announcements).toBe(announcements);
  });

  it("Renders the correct HTML", () => {
    expect(wrapper.find("[data-test='announcement-list']").exists()).toBeTruthy();
    expect(wrapper.find("[data-test='announcement-uuid']").exists()).toBeTruthy();
    expect(wrapper.find("[data-test='announcement-title']").exists()).toBeTruthy();
    expect(wrapper.find("[data-test='announcement-actions']").exists()).toBeTruthy();
  });
});
