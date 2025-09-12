import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import useAnnouncementStore from "@admin/store/modules/announcement";
import AnnouncementDelete from "@admin/components/Announcement/AnnouncementDelete.vue";
import routes from "@admin/router";
import { SnackbarPlugin } from "@/plugins/snackbar";

type AnnouncementDeleteWrapper = VueWrapper<InstanceType<typeof AnnouncementDelete>>;

describe("Announcement Delete", () => {
  const vuetify = createVuetify();

  let wrapper: AnnouncementDeleteWrapper;

  beforeEach(() => {
    setActivePinia(createPinia());
    const announcementStore = useAnnouncementStore();

    announcementStore.announcement = {
      uuid: "eac7e18d-7127-41ca-b68b-8242dfdbaf4c",
      title: "Announcement 1",
      content: "## ShellHub new features \n - New feature 1 \n - New feature 2 \n - New feature 3",
      date: "2022-12-15T19:45:45.618Z",
    };

    wrapper = mount(AnnouncementDelete, {
      global: {
        plugins: [vuetify, routes, SnackbarPlugin],
      },
      props: {
        uuid: "eac7e18d-7127-41ca-b68b-8242dfdbaf4c",
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.exists()).toBe(true);
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Opens the dialog on delete button click", async () => {
    expect(wrapper.vm.showDialog).toBe(false);

    const deleteButton = wrapper.find('[data-test="delete-button"]');
    await deleteButton.trigger("click");

    expect(wrapper.vm.showDialog).toBe(true);
  });
});
