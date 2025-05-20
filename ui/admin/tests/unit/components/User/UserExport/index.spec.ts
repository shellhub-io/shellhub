import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import useUsersStore from "@admin/store/modules/users";
import { SnackbarPlugin } from "@/plugins/snackbar";
import UserExport from "../../../../../src/components/User/UserExport.vue";

type UserExportWrapper = VueWrapper<InstanceType<typeof UserExport>>;

describe("User Export", () => {
  let wrapper: UserExportWrapper;

  beforeEach(() => {
    setActivePinia(createPinia());
    const vuetify = createVuetify();

    const userStore = useUsersStore();

    vi.spyOn(userStore, "setFilterUsers").mockResolvedValue(undefined);
    vi.spyOn(userStore, "exportUsersToCsv").mockResolvedValue("fake-csv-data");

    wrapper = mount(UserExport, {
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

  it("Has default values set correctly", () => {
    expect(wrapper.vm.gtNumberOfNamespaces).toBe(0);
    expect(wrapper.vm.eqNumberOfNamespaces).toBe(0);
    expect(wrapper.vm.dialog).toBe(false);
    expect(wrapper.vm.selected).toBe("moreThan");
  });
});
