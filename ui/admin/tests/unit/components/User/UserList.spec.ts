import { createVuetify } from "vuetify";
import { mount } from "@vue/test-utils";
import { describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import useUsersStore from "@admin/store/modules/users";
import UserList from "@admin/components/User/UserList.vue";
import routes from "@admin/router";
import { SnackbarPlugin } from "@/plugins/snackbar";
import { IAdminUser } from "@admin/interfaces/IUser";

const mockUsers: IAdminUser[] = [
  {
    status: "confirmed" as const,
    created_at: "2022-04-13T11:42:49.578Z",
    email: "depaulacostaantony@gmail.com",
    recovery_email: "blabla@gmail.com",
    mfa: { enabled: false },
    id: "6256b739302b50b6cc5eafcc",
    last_login: "0001-01-01T00:00:00Z",
    name: "antony",
    namespacesOwned: 2,
    max_namespaces: 5,
    username: "antony",
    preferences: {
      auth_methods: ["saml" as const],
    },
  },
];

describe("UserList", async () => {
  setActivePinia(createPinia());
  const vuetify = createVuetify();

  const usersStore = useUsersStore();

  usersStore.users = mockUsers;
  usersStore.usersCount = mockUsers.length;
  usersStore.fetchUsersList = vi.fn();

  const wrapper = mount(UserList, {
    global: {
      plugins: [vuetify, routes, SnackbarPlugin],
    },
  });

  await wrapper.vm.$nextTick();

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
  });
});
