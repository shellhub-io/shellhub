import { createStore } from "vuex";
import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { key } from "../../../../src/store";
import routes from "../../../../src/router";
import NewAnnouncement from "../../../../src/views/NewAnnouncement.vue";

describe("New Announcement", () => {
  const store = createStore({
    state: {
    },
    getters: {
    },
    actions: {
      "announcement/postAnnouncement": vi.fn(),
      "announcement/getAnnouncements": vi.fn(),
      "announcement/announcements": vi.fn(),
      "snackbar/showSnackbarErrorAction": vi.fn(),
    },
  });

  const vuetify = createVuetify();

  let wrapper: VueWrapper<any>;

  beforeEach(() => {
    wrapper = mount(NewAnnouncement, {
      global: {
        plugins: [[store, key], vuetify, routes],
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

  it("Renders the error message when the Title are empty", async () => {
    await wrapper.find("[data-test='announcement-btn-post']").trigger("click");
    expect(wrapper.vm.titleError).toBeTruthy();
  });

  it("Renders the error message when the announcement are empty", async () => {
    wrapper.vm.title = "News ShellHub";
    await wrapper.find("[data-test='announcement-btn-post']").trigger("click");
    expect(wrapper.find("[data-test='announcement-error']").exists()).toBeTruthy();
  });
});
