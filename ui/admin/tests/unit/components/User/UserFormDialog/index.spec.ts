import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import useUsersStore from "@admin/store/modules/users";
import UserFormDialog from "@admin/components/User/UserFormDialog.vue";
import { IUser } from "@admin/interfaces/IUser";
import { SnackbarPlugin } from "@/plugins/snackbar";

type UserFormDialogWrapper = VueWrapper<InstanceType<typeof UserFormDialog>>;

const user: IUser = {
  id: "5f1996c84d2190a22d5857bb",
  name: "Antony",
  email: "antony@gmail.com",
  username: "antony",
  password: "123456789",
  status: "confirmed",
  namespaces: 1,
  max_namespaces: 10,
  created_at: "2023-10-01T12:00:00Z",
  last_login: "2023-10-01T12:00:00Z",
  preferences: {
    auth_methods: ["saml", "local"],
  },
};

describe("UserFormDialog With prop 'createUser' equals false", () => {
  let wrapper: UserFormDialogWrapper;

  beforeEach(() => {
    setActivePinia(createPinia());
    const vuetify = createVuetify();

    const userStore = useUsersStore();

    vi.spyOn(userStore, "put").mockResolvedValue(undefined);
    vi.spyOn(userStore, "refresh").mockResolvedValue(undefined);

    wrapper = mount(UserFormDialog, {
      props: {
        titleCard: "Edit User",
        createUser: false,
        user,
      },
      global: {
        plugins: [vuetify, SnackbarPlugin],
      },
    });

    wrapper.vm.showDialog = true;
  });

  it("Is a Vue instance", () => {
    expect(wrapper.exists()).toBe(true);
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Compare data with default value", () => {
    expect(wrapper.vm.titleCard).toEqual("Edit User");
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

describe("UserFormDialog With prop 'createUser' equals true", () => {
  let wrapper: UserFormDialogWrapper;

  beforeEach(() => {
    setActivePinia(createPinia());
    const vuetify = createVuetify();

    wrapper = mount(UserFormDialog, {
      props: {
        titleCard: "Add User",
        createUser: true,
      },
      global: {
        plugins: [vuetify, SnackbarPlugin],
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.exists()).toBe(true);
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Compare data with default value", () => {
    expect(wrapper.vm.titleCard).toEqual("Add User");
    expect(wrapper.vm.createUser).toEqual(true);
    expect(wrapper.vm.isConfirmed).toBe(false);
  });
});
