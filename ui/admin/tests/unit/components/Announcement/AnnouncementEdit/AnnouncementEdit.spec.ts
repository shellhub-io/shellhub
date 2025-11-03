import { createVuetify } from "vuetify";
import { DOMWrapper, mount } from "@vue/test-utils";
import { describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import useAnnouncementStore from "@admin/store/modules/announcement";
import AnnouncementEdit from "@admin/components/Announcement/AnnouncementEdit.vue";
import { SnackbarPlugin } from "@/plugins/snackbar";

vi.mock("@tinymce/tinymce-vue", () => ({
  default: {
    name: "Editor",
    template: '<div class="tinymce-mock"><textarea v-model="modelValue"></textarea></div>',
    props: ["modelValue", "init", "apiKey"],
  },
}));

const shortAnnouncement = {
  uuid: "eac7e18d-7127-41ca-b68b-8242dfdbaf4c",
  title: "Announcement 1",
  date: "2022-12-15T19:45:45.618Z",
};

const announcement = {
  ...shortAnnouncement,
  content: "## ShellHub new features \n - New feature 1 \n - New feature 2 \n - New feature 3",
};

describe("Announcement Edit", () => {
  setActivePinia(createPinia());
  const announcementStore = useAnnouncementStore();

  announcementStore.announcement = announcement;
  announcementStore.fetchAnnouncement = vi.fn().mockResolvedValue(announcement);

  const wrapper = mount(AnnouncementEdit, {
    global: { plugins: [createVuetify(), SnackbarPlugin] },
    props: { announcementItem: shortAnnouncement },
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders the correct data", () => {
    expect(wrapper.vm.showDialog).toBe(false);
    expect(wrapper.vm.announcement.uuid).toBe(shortAnnouncement.uuid);
    expect(wrapper.vm.announcement.title).toBe(shortAnnouncement.title);
    expect(wrapper.vm.announcement.date).toBe(shortAnnouncement.date);
    expect(wrapper.vm.contentInHtml).toBe("");
    expect(wrapper.vm.contentError).toBe(false);
    expect(wrapper.vm.title).toBe(announcement.title);
  });

  it("Opens the dialog on edit button click", async () => {
    expect(wrapper.vm.showDialog).toBe(false);

    const editButton = wrapper.find('[data-test="edit-button"]');
    await editButton.trigger("click");

    expect(wrapper.vm.showDialog).toBe(true);
    const dialog = new DOMWrapper(document.body);
    expect(dialog.html()).toMatchSnapshot();
  });
});
