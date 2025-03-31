import { createVuetify } from "vuetify";
import { createStore } from "vuex";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import UserFormDialog from "@admin/components/User/UserFormDialog.vue";
import { VNumberInput } from "vuetify/labs/VNumberInput";
import { key } from "../../../../../src/store";

type UserFormDialogWrapper = VueWrapper<InstanceType<typeof UserFormDialog>>;

const store = createStore({
  state: {},
  getters: {},
  actions: {
    "users/remove": () => vi.fn(),
    "snackbar/showSnackbarSuccessAction": vi.fn(),
    "snackbar/showSnackbarErrorAction": vi.fn(),
  },
});

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
    const vuetify = createVuetify({
      components: {
        VNumberInput,
      },
    });

    wrapper = mount(UserFormDialog, {
      props: {
        titleCard: "Edit User",
        createUser: false,
        user,
      },
      global: {
        plugins: [[store, key], vuetify],
      },
    });

    // Ensure the dialog is open to populate form fields
    wrapper.vm.openDialog();
  });

  it("Is a Vue instance", () => {
    expect(wrapper).toBeTruthy();
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
    expect(wrapper.vm.password).toEqual(undefined);
    expect(wrapper.vm.userConfirmed).toEqual(user.confirmed);
  });
});

describe("UserFormDialog With prop 'createUser' equals true", () => {
  let wrapper: UserFormDialogWrapper;

  beforeEach(() => {
    const vuetify = createVuetify({
      components: {
        VNumberInput,
      },
    });

    wrapper = mount(UserFormDialog, {
      props: {
        titleCard: "Add User",
        createUser: true,
      },
      global: {
        plugins: [[store, key], vuetify],
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Compare data with default value", () => {
    expect(wrapper.vm.titleCard).toEqual("Add User");
    expect(wrapper.vm.createUser).toEqual(true);
    expect(wrapper.vm.emailIsConfirmed).toEqual(undefined);
  });
});
