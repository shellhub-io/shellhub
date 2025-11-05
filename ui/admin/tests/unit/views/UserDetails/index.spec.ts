import { createVuetify } from "vuetify";
import { mount } from "@vue/test-utils";
import { describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import { createMemoryHistory, createRouter } from "vue-router";
import useUsersStore from "@admin/store/modules/users";
import useAuthStore from "@admin/store/modules/auth";
import UserDetails from "@admin/views/UserDetails.vue";
import { SnackbarPlugin } from "@/plugins/snackbar";

const user = {
  status: "confirmed",
  created_at: "2022-04-13T11:42:49.578Z",
  email: "antony@gmail.com",
  id: "6256b739302b50b6cc5eafcc",
  last_login: "0001-01-01T00:00:00Z",
  name: "antony",
  username: "antony",
  namespacesOwned: 2,
};

describe("User Details", async () => {
  const pinia = createPinia();
  setActivePinia(pinia);

  const usersStore = useUsersStore();
  usersStore.fetchUserById = vi.fn().mockResolvedValue(user);

  const authStore = useAuthStore();
  authStore.getLoginToken = vi.fn().mockResolvedValue("mock-token");

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

  const wrapper = mount(UserDetails, {
    global: {
      plugins: [pinia, vuetify, router, SnackbarPlugin],
    },
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
  });

  it("Should render the correct buttons", () => {
    expect(wrapper.find("a.v-icon.mdi-login").exists()).toBe(true);
    expect(wrapper.find("a.v-icon.mdi-delete").exists()).toBe(true);
  });
});
