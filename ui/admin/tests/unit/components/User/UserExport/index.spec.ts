import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import useUsersStore from "@admin/store/modules/users";
import UserExport from "@admin/components/User/UserExport.vue";
import { SnackbarPlugin } from "@/plugins/snackbar";

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
    expect(wrapper.vm.numberOfNamespaces).toBe(0);
    expect(wrapper.vm.showDialog).toBe(false);
    expect(wrapper.vm.selectedFilter).toBe("moreThan");
  });

  it("Renders the dialog when showDialog is true", async () => {
    wrapper.vm.showDialog = true;
    await wrapper.vm.$nextTick();
    expect(wrapper.findComponent({ name: "VDialog" }).exists()).toBe(true);
  });
});
