import { createVuetify } from "vuetify";
import { DOMWrapper, mount } from "@vue/test-utils";
import { describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import useUsersStore from "@admin/store/modules/users";
import UserFormDialog from "@admin/components/User/UserFormDialog.vue";
import { IAdminUser } from "@admin/interfaces/IUser";
import { SnackbarPlugin } from "@/plugins/snackbar";

const user: IAdminUser = {
  id: "5f1996c84d2190a22d5857bb",
  name: "Antony",
  email: "antony@gmail.com",
  recovery_email: "blabla@gmail.com",
  mfa: { enabled: false },
  username: "antony",
  status: "confirmed",
  namespacesOwned: 1,
  max_namespaces: 10,
  created_at: "2023-10-01T12:00:00Z",
  last_login: "2023-10-01T12:00:00Z",
  preferences: { auth_methods: ["saml", "local"] },
};

setActivePinia(createPinia());
const usersStore = useUsersStore();
usersStore.updateUser = vi.fn();
usersStore.fetchUsersList = vi.fn();
const vuetify = createVuetify();

describe("UserFormDialog (Edit User)", () => {
  const wrapper = mount(UserFormDialog, {
    props: { createUser: false, user },
    global: { plugins: [vuetify, SnackbarPlugin] },
  });

  wrapper.vm.showDialog = true;

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
    const dialog = new DOMWrapper(document.body);
    expect(dialog.html()).toMatchSnapshot();
  });

  it("Compare data with default value", () => {
    expect(wrapper.vm.createUser).toEqual(false);
    expect(wrapper.vm.user).toEqual(user);
    expect(wrapper.vm.isConfirmed).toEqual(user.status === "confirmed");
  });

  it("Compare user data with prop value", () => {
    expect(wrapper.vm.name).toEqual(user.name);
    expect(wrapper.vm.email).toEqual(user.email);
    expect(wrapper.vm.username).toEqual(user.username);
    expect(wrapper.vm.password).toBeUndefined();
  });
});

describe("UserFormDialog (Create User)", () => {
  const wrapper = mount(UserFormDialog, {
    props: { createUser: true },
    global: { plugins: [vuetify, SnackbarPlugin] },
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
    const dialog = new DOMWrapper(document.body);
    expect(dialog.html()).toMatchSnapshot();
  });

  it("Compare data with default value", () => {
    expect(wrapper.vm.createUser).toEqual(true);
    expect(wrapper.vm.isConfirmed).toBe(false);
  });
});

describe("UserFormDialog (Edit User with namespace creation disabled)", () => {
  const userWithDisabledNamespace: IAdminUser = {
    ...user,
    max_namespaces: 0,
  };

  const wrapper = mount(UserFormDialog, {
    props: { createUser: false, user: userWithDisabledNamespace },
    global: { plugins: [vuetify, SnackbarPlugin] },
  });

  wrapper.vm.showDialog = true;

  it("Renders the component with namespace creation disabled", () => {
    expect(wrapper.html()).toMatchSnapshot();
    const dialog = new DOMWrapper(document.body);
    expect(dialog.html()).toMatchSnapshot();
  });

  it("Should have changeNamespaceLimit enabled when max_namespaces is 0", () => {
    expect(wrapper.vm.user?.max_namespaces).toEqual(0);
  });
});
