import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import useAnnouncementStore from "@admin/store/modules/announcement";
import { SnackbarPlugin } from "@/plugins/snackbar";
import routes from "../../../../src/router";
import NewAnnouncement from "../../../../src/views/NewAnnouncement.vue";

type NewAnnouncementWrapper = VueWrapper<InstanceType<typeof NewAnnouncement>>;

describe("New Announcement", () => {
  let wrapper: NewAnnouncementWrapper;

  beforeEach(() => {
    const pinia = createPinia();
    setActivePinia(pinia);

    const announcementStore = useAnnouncementStore();

    announcementStore.postAnnouncement = vi.fn();
    announcementStore.fetchAnnouncements = vi.fn();

    const vuetify = createVuetify();

    wrapper = mount(NewAnnouncement, {
      global: {
        plugins: [pinia, vuetify, routes, SnackbarPlugin],
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper).toBeTruthy();
  });

  it("Renders the component", () => {
    const html = wrapper.html().replace(/id="tiny-vue_\d+"/g, 'id="tiny-vue_random"');
    expect(html).toMatchSnapshot();
  });

  it("Has the correct title", () => {
    expect(wrapper.find("h1").text()).toBe("Create new Announcement");
  });

  it("Renders the correct HTML", () => {
    expect(wrapper.find("[data-test='announcement-title']").exists()).toBeTruthy();
    expect(wrapper.find("[data-test='announcement-content']").exists()).toBeTruthy();
    expect(wrapper.find("[data-test='announcement-error']").exists()).toBeFalsy();
    expect(wrapper.find("[data-test='announcement-btn-post']").exists()).toBeTruthy();
  });

  it("Renders the error message when the Title is empty", async () => {
    await wrapper.find("[data-test='announcement-btn-post']").trigger("click");
    expect(wrapper.vm.titleError).toBeTruthy();
  });

  it("Renders the error message when the announcement is empty", async () => {
    wrapper.vm.title = "News ShellHub";
    await wrapper.find("[data-test='announcement-btn-post']").trigger("click");
    expect(wrapper.find("[data-test='announcement-error']").exists()).toBeTruthy();
  });
});
