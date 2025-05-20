import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import useUsersStore from "@admin/store/modules/users";
import useAuthStore from "@admin/store/modules/auth";
import useSnackbarStore from "@admin/store/modules/snackbar";
import { createMemoryHistory, createRouter } from "vue-router";
import { SnackbarPlugin } from "@/plugins/snackbar";
import UserDetails from "../../../../src/views/UserDetails.vue";

type UserDetailsWrapper = VueWrapper<InstanceType<typeof UserDetails>>;

const user = {
  confirmed: true,
  created_at: "2022-04-13T11:42:49.578Z",
  email: "antony@gmail.com",
  id: "6256b739302b50b6cc5eafcc",
  last_login: "0001-01-01T00:00:00Z",
  name: "antony",
  namespaces: 0,
  password: "somepassword",
  username: "antony",
};

describe("User Details", () => {
  let wrapper: UserDetailsWrapper;

  beforeEach(async () => {
    const pinia = createPinia();
    setActivePinia(pinia);

    const usersStore = useUsersStore();
    usersStore.get = vi.fn().mockResolvedValue(undefined);
    usersStore.user = user;

    const authStore = useAuthStore();
    authStore.loginToken = vi.fn().mockResolvedValue("mock-token");

    const snackbarStore = useSnackbarStore();
    snackbarStore.showSnackbarErrorAction = vi.fn();

    const vuetify = createVuetify();

    const router = createRouter({
      history: createMemoryHistory(),
      routes: [
        {
          path: "/user/:id",
          name: "userDetails",
          component: UserDetails,
        },
      ],
    });

    await router.push({ name: "userDetails", params: { id: user.id } });
    await router.isReady();

    wrapper = mount(UserDetails, {
      global: {
        plugins: [pinia, vuetify, router, SnackbarPlugin],
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
    expect(wrapper.vm.currentUser).toEqual(user);
  });

  it("Render the correct title", () => {
    expect(wrapper.find("h1").text()).toEqual("User Details");
  });

  it("Should render the props of the user in the screen", () => {
    expect(wrapper.find(`[data-test='${user.id}']`).text()).toContain(user.id);
    expect(wrapper.find(`[data-test='${user.email}']`).text()).toContain(user.email);
    expect(wrapper.find(`[data-test='${user.username}']`).text()).toContain(user.username);
    expect(wrapper.find(`[data-test='${user.namespaces}']`).text()).toContain(`${user.namespaces}`);
  });

  it("Should render the correct buttons", () => {
    expect(wrapper.find("a.v-icon.mdi-login").exists()).toBe(true);
    expect(wrapper.find("a.v-icon.mdi-delete").exists()).toBe(true);
  });
});
