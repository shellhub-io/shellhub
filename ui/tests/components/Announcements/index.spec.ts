import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it } from "vitest";
import Announcements from "../../../src/components/Announcements/AnnouncementsModal.vue";

const announcement = {
  uuid: "52088548-2b99-4f38-ac09-3a8f8988476f",
  title: "This is a announcement",
  content: "## ShellHub new features \n - New feature 1 \n - New feature 2 \n - New feature 3",
  date: "2022-12-15T19:45:45.618Z"
};

const announcementContentInHtml = `<h2>ShellHub new features</h2>
<ul>
<li>New feature 1</li>
<li>New feature 2</li>
<li>New feature 3</li>
</ul>
`;

describe("Announcements", () => {
  let wrapper: VueWrapper<any>;

  beforeEach(async () => {
    const vuetify = createVuetify();
    wrapper = mount(Announcements, {
      props: {
        show: true,
        announcement: announcement,
      },
      global: {
        plugins: [vuetify],
      },
    });    
  });

  it("Is a Vue instance", () => {
    expect(wrapper).toBeTruthy();
  });
  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  ///////
  // Data checking
  //////
  it("Has the correct props", () => {
    expect(wrapper.vm.show).toBe(true);
    expect(wrapper.vm.announcement).toStrictEqual(announcement);
  });

  it("Has the correct data", () => {
    expect(wrapper.vm.date).toBe("December 15, 2022");
    expect(wrapper.vm.markdownContent).toBe(announcementContentInHtml);
  });
});