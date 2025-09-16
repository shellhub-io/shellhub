import { createVuetify } from "vuetify";
import { mount } from "@vue/test-utils";
import { describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import useUsersStore from "@admin/store/modules/users";
import UserDelete from "@admin/components/User/UserDelete.vue";
import routes from "@admin/router";
import { SnackbarPlugin } from "@/plugins/snackbar";

describe("User Delete", () => {
  setActivePinia(createPinia());
  const usersStore = useUsersStore();
  const vuetify = createVuetify();

  usersStore.deleteUser = vi.fn();
  usersStore.fetchUsersList = vi.fn();

  const wrapper = mount(UserDelete, {
    props: {
      id: "6256d9e3ea6f26bc595130fa",
      redirect: false,
    },
    global: {
      plugins: [vuetify, routes, SnackbarPlugin],
    },
  });

  it("Is a Vue instance", () => {
    expect(wrapper.exists()).toBe(true);
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Receives props correctly", () => {
    expect(wrapper.vm.id).toBe("6256d9e3ea6f26bc595130fa");
    expect(wrapper.vm.redirect).toBe(false);
  });

  it("Dialog should be false by default", () => {
    expect(wrapper.vm.showDialog).toBe(false);
  });
});
