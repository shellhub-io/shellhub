import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import useUsersStore from "@admin/store/modules/users";
import useSnackbarStore from "@admin/store/modules/snackbar";
import { INotificationsError, INotificationsSuccess } from "@admin/interfaces/INotifications";
import { SnackbarPlugin } from "@/plugins/snackbar";
import routes from "../../../../../src/router";
import UserDelete from "../../../../../src/components/User/UserDelete.vue";

type UserDeleteWrapper = VueWrapper<InstanceType<typeof UserDelete>>;

describe("User Delete", () => {
  let wrapper: UserDeleteWrapper;

  beforeEach(() => {
    setActivePinia(createPinia());

    const vuetify = createVuetify();

    const usersStore = useUsersStore();
    const snackbarStore = useSnackbarStore();

    vi.spyOn(usersStore, "remove").mockResolvedValue(undefined);
    vi.spyOn(usersStore, "refresh").mockResolvedValue(undefined);
    vi.spyOn(snackbarStore, "showSnackbarSuccessAction").mockImplementation(() => INotificationsSuccess.userDelete);
    vi.spyOn(snackbarStore, "showSnackbarErrorAction").mockImplementation(() => INotificationsError.userDelete);

    wrapper = mount(UserDelete, {
      props: {
        id: "6256d9e3ea6f26bc595130fa",
        redirect: false,
      },
      global: {
        plugins: [vuetify, routes, SnackbarPlugin],
      },
    });
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
    expect(wrapper.vm.dialog).toBe(false);
  });
});
