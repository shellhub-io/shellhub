import { createStore } from "vuex";
import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { key } from "../../../../src/store";
import routes from "../../../../src/router";
import Users from "../../../../src/views/Users.vue";

describe("Users", () => {
  const store = createStore({
    state: {},
    getters: {
      "users/perPage": () => 10,
      "users/page": () => 1,
      "users/numberUsers": () => 1,
    },
    actions: {
      "users/search": vi.fn(),
      "users/fetch": vi.fn(),
      "snackbar/showSnackbarErrorAction": vi.fn(),
    },
  });
  let wrapper: VueWrapper<any>;

  beforeEach(() => {
    const vuetify = createVuetify();

    wrapper = mount(Users, {
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

  it("Renders the template with default data", async () => {
    expect(wrapper.vm.filter).toBe("");
  });

  it("Must change the filter value when input change", async () => {
    expect(wrapper.vm.filter).toEqual("");
    await wrapper.find("input").setValue("ShellHub");
    expect(wrapper.vm.filter).toEqual("ShellHub");
  });

  it("Should render all the components in the screen", () => {
    expect(wrapper.find("h1").text()).toContain("Users");
    expect(wrapper.find("[data-test='users-list']").exists()).toBe(true);
    expect(wrapper.find("[data-test='users-export-btn']").exists()).toBe(true);
    expect(wrapper.find("[data-test='user-add-btn']").exists()).toBe(true);
    expect(wrapper.find("input").exists()).toBe(true);
  });
});
