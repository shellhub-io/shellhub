import { createStore } from "vuex";
import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { key } from "../../../../src/store";
import routes from "../../../../src/router";
import UserDetails from "../../../../src/views/UserDetails.vue";

const user = {
  confirmed: true,
  created_at: "2022-04-13T11:42:49.578Z",
  email: "antony@gmail.com",
  id: "6256b739302b50b6cc5eafcc",
  last_login: "0001-01-01T00:00:00Z",
  name: "antony",
  namespaces: 0,
  password: "15e2b0d3c33891ebb0f1ef609ec419420c20e320ce94c65fbc8c3312448eb225",
  username: "antony",
};

const mockRoute = {
  params: {
    id: "6256b739302b50b6cc5eafcc",
  },
};

describe("User Details", () => {
  const store = createStore({
    state: {},
    getters: {
      "users/user": () => user,
    },
    actions: {
      "users/get": vi.fn(),
      "auth/loginToken": vi.fn(),
      "snackbar/showSnackbarErrorAction": vi.fn(),
    },
  });
  let wrapper: VueWrapper<any>;

  beforeEach(() => {
    const vuetify = createVuetify();

    wrapper = mount(UserDetails, {
      global: {
        plugins: [[store, key], vuetify, routes],
        mocks: {
          $route: mockRoute,
        },
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Has the correct data", async () => {
    expect(wrapper.vm.currentUser).toEqual(user);
  });

  it("Render the correct title", () => {
    expect(wrapper.find("h1").text()).toEqual("User Details");
  });

  it("Should render the props of the user in the Screen", () => {
    expect(wrapper.find("[data-test='antony']").text()).toContain(user.name);
    expect(wrapper.find("[data-test='antony@gmail.com']").text()).toContain(user.email);
    expect(wrapper.find("[data-test='antony']").text()).toContain(user.username);
    expect(wrapper.find("[data-test='0']").text()).toContain(user.namespaces);
  });

  it("Should render the correct buttons", () => {
    expect(wrapper.find("a.v-icon.mdi-login").exists()).toBe(true);
    expect(wrapper.find("a.v-icon.mdi-delete").exists()).toBe(true);
  });
});
