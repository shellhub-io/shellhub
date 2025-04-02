import { createStore } from "vuex";
import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { key } from "../../../../src/store";
import routes from "../../../../src/router";
import AnnouncementDetails from "../../../../src/views/AnnouncementDetails.vue";

type AnnouncementDetailsWrapper = VueWrapper<InstanceType<typeof AnnouncementDetails>>;

const announcementDetail = {
  uuid: "eac7e18d-7127-41ca-b68b-8242dfdbaf4c",
  title: "Announcement 1",
  content: "## ShellHub new features \n - New feature 1 \n - New feature 2 \n - New feature 3",
  date: "2022-12-15T19:45:45.618Z",
};

const announcementContentInHtml = `<h2>ShellHub new features</h2>
<ul>
<li>New feature 1</li>
<li>New feature 2</li>
<li>New feature 3</li>
</ul>
`;

const mockRoute = {
  params: {
    uuid: "eac7e18d-7127-41ca-b68b-8242dfdbaf4c",
  },
};

describe("Announcement Details", () => {
  const store = createStore({
    state: {
      announcement: announcementDetail,
    },
    getters: {
      "announcement/announcement": () => announcementDetail,
    },
    actions: {
      "announcement/getAnnouncement": vi.fn(),
      "announcement/announcements": vi.fn(),
    },
  });

  const vuetify = createVuetify();

  let wrapper: AnnouncementDetailsWrapper;

  beforeEach(() => {
    wrapper = mount(AnnouncementDetails, {
      global: {
        plugins: [[store, key], vuetify, routes],
        mocks: {
          $route: mockRoute,
          $router: {
            push: vi.fn(),
          },
        },
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Has the correct data", async () => {
    expect(wrapper.vm.announcement).toEqual(announcementDetail);
  });

  it("Render the correct title", () => {
    expect(wrapper.find("h1").text()).toEqual("Announcement Details");
  });

  it("Should render the props of the announcement in the Screen", () => {
    expect(wrapper.find(`[data-test='${announcementDetail.uuid}']`).text()).toContain(announcementDetail.uuid);
    expect(wrapper.find(`[data-test='${announcementDetail.title}']`).text()).toContain(announcementDetail.title);
  });

  it("Shoud render the markdown content in HTML", () => {
    expect(wrapper.vm.contentToHtml).toBe(announcementContentInHtml);
  });

  it("Should render the date in the correct format", () => {
    expect(wrapper.vm.date).toBe("December 15, 2022");
  });
});
