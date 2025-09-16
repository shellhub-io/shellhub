import { createVuetify } from "vuetify";
import { mount } from "@vue/test-utils";
import { describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import useUsersStore from "@admin/store/modules/users";
import routes from "@admin/router";
import Users from "@admin/views/Users.vue";
import { SnackbarPlugin } from "@/plugins/snackbar";

describe("Users", () => {
  const pinia = createPinia();
  setActivePinia(pinia);
  const usersStore = useUsersStore();
  usersStore.fetchUsersList = vi.fn();

  const vuetify = createVuetify();

  const wrapper = mount(Users, {
    global: {
      plugins: [pinia, vuetify, routes, SnackbarPlugin],
    },
  });

  it("Is a Vue instance", () => {
    expect(wrapper).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders the template with default data", () => {
    expect(wrapper.vm.filter).toBe("");
  });

  it("Must change the filter value when input change", async () => {
    expect(wrapper.vm.filter).toBe("");
    const input = wrapper.find("input");
    await input.setValue("ShellHub");
    expect(wrapper.vm.filter).toBe("ShellHub");
  });

  it("Should render all the components in the screen", () => {
    expect(wrapper.find("h1").text()).toContain("Users");
    expect(wrapper.find("[data-test='users-list']").exists()).toBe(true);
    expect(wrapper.find("[data-test='users-export-btn']").exists()).toBe(true);
    expect(wrapper.find("[data-test='user-add-btn']").exists()).toBe(true);
    expect(wrapper.find("input").exists()).toBe(true);
  });
});
