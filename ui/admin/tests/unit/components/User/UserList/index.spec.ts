import { createVuetify } from "vuetify";
import { createStore } from "vuex";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { VNumberInput } from "vuetify/labs/VNumberInput";
import UserList from "../../../../../src/components/User/UserList.vue";
import { key } from "../../../../../src/store";
import routes from "../../../../../src/router";

type UserListWrapper = VueWrapper<InstanceType<typeof UserList>>;

const users = [
  {
    confirmed: true,
    created_at: "2022-04-13T11:42:49.578Z",
    email: "depaulacostaantony@gmail.com",
    id: "6256b739302b50b6cc5eafcc",
    last_login: "0001-01-01T00:00:00Z",
    name: "antony",
    namespaces: 2,
    password: "15e2b0d3c33891ebb0f1ef609ec419420c20e320ce94c65fbc8c3312448eb225",
    username: "antony",
  },
];

const store = createStore({
  state: {
    users,
  },
  getters: {
    "users/users": (state) => state.users,
    "users/numberUsers": (state) => state.users.length,
  },
  actions: {
    "snackbar/showSnackbarErrorAction": vi.fn(),
    "snackbar/setSnackbarErrorDefault": vi.fn(),
    "users/refresh": vi.fn(),
    "users/fetch": vi.fn(),
  },
});

describe("UserList", () => {
  let wrapper: UserListWrapper;

  beforeEach(() => {
    const vuetify = createVuetify({
      components: {
        VNumberInput,
      },
    });

    wrapper = mount(UserList, {
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

  it("Renders data in the computed", async () => {
    const componentUsers = await wrapper.vm.users;
    expect(componentUsers).toEqual(users);
  });

  it("Should render the props of the user in the table", () => {
    expect(wrapper.find("[name-test]").text()).toContain(users[0].name);
    expect(wrapper.find("[email-test]").text()).toContain(users[0].email);
    expect(wrapper.find("[username-test]").text()).toContain(users[0].username);
    expect(wrapper.find("[namespaces-test]").text()).toContain(users[0].namespaces);
  });
});
