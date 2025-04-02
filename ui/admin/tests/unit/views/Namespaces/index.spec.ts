import { createStore } from "vuex";
import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { key } from "../../../../src/store";
import routes from "../../../../src/router";
import Namespaces from "../../../../src/views/Namespaces.vue";

type NamespacesWrapper = VueWrapper<InstanceType<typeof Namespaces>>;

describe("Namespaces", () => {
  const store = createStore({
    state: {},
    getters: {
      "namespaces/perPage": () => 10,
      "namespaces/page": () => 1,
      "namespaces/numberOfNamespaces": () => 1,
    },
    actions: {
      "namespaces/search": vi.fn(),
      "namespaces/fetch": vi.fn(),
      "snackbar/showSnackbarErrorAction": vi.fn(),
    },
  });
  let wrapper: NamespacesWrapper;

  beforeEach(() => {
    const vuetify = createVuetify();

    wrapper = mount(Namespaces, {
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
    expect(wrapper.find("h1").text()).toContain("Namespaces");
    expect(wrapper.find("[data-test='namespaces-list']").exists()).toBe(true);
    expect(wrapper.find("[data-test='namespaces-export-btn']").exists()).toBe(true);
    expect(wrapper.find("input").exists()).toBe(true);
  });
});
