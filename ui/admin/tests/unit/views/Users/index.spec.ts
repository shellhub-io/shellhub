import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import useUsersStore from "@admin/store/modules/users";
import { SnackbarPlugin } from "@/plugins/snackbar";
import routes from "../../../../src/router";
import Users from "../../../../src/views/Users.vue";

type UsersWrapper = VueWrapper<InstanceType<typeof Users>>;

describe("Users", () => {
  let wrapper: UsersWrapper;

  beforeEach(() => {
    const pinia = createPinia();
    setActivePinia(pinia);

    const userStore = useUsersStore();
    const snackbarStore = useSnackbarStore();

    vi.spyOn(userStore, "getPerPage", "get").mockReturnValue(10);
    vi.spyOn(userStore, "getPage", "get").mockReturnValue(1);
    vi.spyOn(userStore, "getNumberUsers", "get").mockReturnValue(1);

    userStore.search = vi.fn();
    userStore.fetch = vi.fn();
    snackbarStore.showSnackbarErrorAction = vi.fn();

    const vuetify = createVuetify();

    wrapper = mount(Users, {
      global: {
        plugins: [pinia, vuetify, routes, SnackbarPlugin],
      },
    });
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
