import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import useAnnouncementStore from "@admin/store/modules/announcement";
import AnnouncementList from "@admin/components/Announcement/AnnouncementList.vue";
import routes from "@admin/router";
import { SnackbarPlugin } from "@/plugins/snackbar";

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

describe("Announcement List", () => {
  const vuetify = createVuetify();
  let wrapper: AnnouncementListWrapper;

  beforeEach(() => {
    setActivePinia(createPinia());

    const announcementStore = useAnnouncementStore();

    vi.spyOn(announcementStore, "fetchAnnouncementList").mockResolvedValue();
    announcementStore.announcements = announcements;
    announcementStore.announcementCount = 2;

    wrapper = mount(AnnouncementList, {
      global: {
        plugins: [vuetify, routes, SnackbarPlugin],
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
    expect(wrapper.vm.itemsPerPage).toBe(10);
    expect(wrapper.vm.page).toBe(1);
    expect(wrapper.vm.loading).toBe(false);
  });

  it("Renders the correct computed", () => {
    expect(wrapper.vm.announcementCount).toBe(2);
    expect(wrapper.vm.announcements).toEqual(announcements);
  });

  it("Renders the correct HTML", () => {
    expect(wrapper.find("[data-test='announcement-list']").exists()).toBeTruthy();
    expect(wrapper.find("[data-test='announcement-uuid']").exists()).toBeTruthy();
    expect(wrapper.find("[data-test='announcement-title']").exists()).toBeTruthy();
    expect(wrapper.find("[data-test='announcement-actions']").exists()).toBeTruthy();
  });
});
