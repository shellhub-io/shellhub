import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import useUsersStore from "@admin/store/modules/users";
import UserFormDialog from "@admin/components/User/UserFormDialog.vue";
import { SnackbarPlugin } from "@/plugins/snackbar";

type UserFormDialogWrapper = VueWrapper<InstanceType<typeof UserFormDialog>>;

const user = {
  id: "5f1996c84d2190a22d5857bb",
  name: "Antony",
  email: "antony@gmail.com",
  username: "antony",
  password: "123456789",
  confirmed: true,
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

    wrapper.vm.openDialog();
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
    expect(wrapper.vm.emailIsConfirmed).toEqual(user.confirmed);
  });

  it("Compare user data with prop value", () => {
    expect(wrapper.vm.name).toEqual(user.name);
    expect(wrapper.vm.email).toEqual(user.email);
    expect(wrapper.vm.username).toEqual(user.username);
    expect(wrapper.vm.password).toBeUndefined();
    expect(wrapper.vm.userConfirmed).toEqual(user.confirmed);
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
    expect(wrapper.vm.emailIsConfirmed).toBeUndefined();
  });
});
