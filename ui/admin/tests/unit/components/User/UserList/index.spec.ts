import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import useUsersStore from "@admin/store/modules/users";
import { SnackbarPlugin } from "@/plugins/snackbar";
import UserList from "../../../../../src/components/User/UserList.vue";
import routes from "../../../../../src/router";

type UserListWrapper = VueWrapper<InstanceType<typeof UserList>>;

const mockUsers = [
  {
    confirmed: true,
    created_at: "2022-04-13T11:42:49.578Z",
    email: "depaulacostaantony@gmail.com",
    id: "6256b739302b50b6cc5eafcc",
    last_login: "0001-01-01T00:00:00Z",
    name: "antony",
    namespaces: 2,
    password: "dummy",
    username: "antony",
    auth_methods: ["saml"],
  },
];

describe("UserList", () => {
  let wrapper: UserListWrapper;

  beforeEach(async () => {
    setActivePinia(createPinia());
    const vuetify = createVuetify();

    const usersStore = useUsersStore();

    usersStore.users = mockUsers;
    usersStore.numberUsers = mockUsers.length;

    vi.spyOn(usersStore, "fetch").mockResolvedValue(true);
    vi.spyOn(usersStore, "refresh").mockResolvedValue();

    wrapper = mount(UserList, {
      global: {
        plugins: [vuetify, routes, SnackbarPlugin],
      },
    });

    await wrapper.vm.$nextTick();
  });

  it("Is a Vue instance", () => {
    expect(wrapper.exists()).toBe(true);
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders users in computed property", () => {
    expect(wrapper.vm.users).toEqual(mockUsers);
  });

  it("Should render user props correctly in the table", () => {
    expect(wrapper.find("[name-test]").text()).toContain(mockUsers[0].name);
    expect(wrapper.find("[email-test]").text()).toContain(mockUsers[0].email);
    expect(wrapper.find("[username-test]").text()).toContain(mockUsers[0].username);
    expect(wrapper.find("[namespaces-test]").text()).toContain(String(mockUsers[0].namespaces));
  });
});
