import { createVuetify } from "vuetify";
import { DOMWrapper, mount } from "@vue/test-utils";
import { describe, expect, it } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import useAnnouncementStore from "@admin/store/modules/announcement";
import AnnouncementDelete from "@admin/components/Announcement/AnnouncementDelete.vue";
import { SnackbarPlugin } from "@/plugins/snackbar";

describe("Announcement Delete", () => {
  setActivePinia(createPinia());
  const announcementStore = useAnnouncementStore();

  announcementStore.announcement = {
    uuid: "eac7e18d-7127-41ca-b68b-8242dfdbaf4c",
    title: "Announcement 1",
    content: "## ShellHub new features \n - New feature 1 \n - New feature 2 \n - New feature 3",
    date: "2022-12-15T19:45:45.618Z",
  };

  const wrapper = mount(AnnouncementDelete, {
    global: { plugins: [createVuetify(), SnackbarPlugin] },
    props: { uuid: "eac7e18d-7127-41ca-b68b-8242dfdbaf4c" },
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Opens the dialog on delete button click", async () => {
    expect(wrapper.vm.showDialog).toBe(false);

    const deleteButton = wrapper.find('[data-test="delete-button"]');
    await deleteButton.trigger("click");

    const dialog = new DOMWrapper(document.body);
    expect(dialog.html()).toMatchSnapshot();
  });
});
