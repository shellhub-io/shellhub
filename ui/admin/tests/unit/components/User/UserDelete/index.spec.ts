import { createVuetify } from "vuetify";
import { DOMWrapper, mount } from "@vue/test-utils";
import { describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import useUsersStore from "@admin/store/modules/users";
import UserDelete from "@admin/components/User/UserDelete.vue";
import routes from "@admin/router";
import { SnackbarPlugin } from "@/plugins/snackbar";

describe("User Delete", () => {
  setActivePinia(createPinia());
  const usersStore = useUsersStore();
  usersStore.deleteUser = vi.fn();
  usersStore.fetchUsersList = vi.fn();

  const wrapper = mount(UserDelete, {
    props: {
      id: "6256d9e3ea6f26bc595130fa",
      redirect: false,
    },
    global: { plugins: [createVuetify(), routes, SnackbarPlugin] },
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
    const dialog = new DOMWrapper(document.body);
    expect(dialog.html()).toMatchSnapshot();
  });

  it("Receives props correctly", () => {
    expect(wrapper.vm.id).toBe("6256d9e3ea6f26bc595130fa");
    expect(wrapper.vm.redirect).toBe(false);
  });
});
