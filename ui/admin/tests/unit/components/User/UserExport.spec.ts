import { nextTick } from "vue";
import { createVuetify } from "vuetify";
import { DOMWrapper, mount } from "@vue/test-utils";
import { describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import useUsersStore from "@admin/store/modules/users";
import UserExport from "@admin/components/User/UserExport.vue";
import { SnackbarPlugin } from "@/plugins/snackbar";

describe("User Export", () => {
  setActivePinia(createPinia());
  const usersStore = useUsersStore();
  vi.spyOn(usersStore, "exportUsersToCsv").mockResolvedValue(new File([], "users.csv"));

  const wrapper = mount(UserExport, {
    global: {
      plugins: [createVuetify(), SnackbarPlugin],
    },
  });

  it("Renders the component", async () => {
    expect(wrapper.html()).toMatchSnapshot();
    wrapper.vm.showDialog = true;
    await nextTick();
    const dialog = new DOMWrapper(document.body);
    expect(dialog.html()).toMatchSnapshot();
  });

  it("Has default values set correctly", () => {
    expect(wrapper.vm.numberOfNamespaces).toBe(0);
    expect(wrapper.vm.selectedFilter).toBe("moreThan");
  });
});
