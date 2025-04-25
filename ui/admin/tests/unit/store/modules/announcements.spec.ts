import { describe, expect, it, beforeEach } from "vitest";
import { setActivePinia, createPinia } from "pinia";
import useAnnouncementStore from "@admin/store/modules/announcement";

describe("Announcement Store (Pinia)", () => {
  let announcementStore: ReturnType<typeof useAnnouncementStore>;

  const announcements = [
    {
      uuid: "52088548-2b99-4f38-ac09-3a8f8988476f",
      title: "This is a announcement",
      content: "## ShellHub new features \n - New feature 1 \n - New feature 2 \n - New feature 3",
      date: "2022-12-15T19:45:45.618Z",
    },
    {
      uuid: "52188548-2b99-4f38-ac09-3a8f8988476f",
      title: "This is a new announcement",
      content: "## ShellHub new features \n - New feature 1 \n - New feature 2 \n - New feature 3",
      date: "2022-12-15T19:46:45.618Z",
    },
  ];

  const numberAnnouncements = 2;

  const announcement = {
    uuid: "52088548-2b99-4f38-ac09-3a8f8988476f",
    title: "This is a announcement",
    content: "## ShellHub new features \n - New feature 1 \n - New feature 2 \n - New feature 3",
    date: "2022-12-15T19:45:45.618Z",
  };

  beforeEach(() => {
    setActivePinia(createPinia());
    announcementStore = useAnnouncementStore();
  });

  it("Returns default state", () => {
    expect(announcementStore.getAnnouncements).toEqual([]);
    expect(announcementStore.getAnnouncement).toEqual({});
    expect(announcementStore.getPage).toBe(1);
    expect(announcementStore.getPerPage).toBe(10);
    expect(announcementStore.getOrderBy).toBe("asc");
  });

  it("Sets announcements and updates count", () => {
    announcementStore.announcements = announcements;
    announcementStore.numberAnnouncements = numberAnnouncements;
    expect(announcementStore.getAnnouncements).toEqual(announcements);
    expect(announcementStore.getNumberAnnouncements).toBe(numberAnnouncements);
  });

  it("Sets a single announcement", () => {
    announcementStore.announcement = announcement;
    expect(announcementStore.getAnnouncement).toEqual(announcement);
  });

  it("Updates page and perPage values", () => {
    announcementStore.setPageAndPerPage({ page: 2, perPage: 20 });
    expect(announcementStore.getPage).toBe(2);
    expect(announcementStore.getPerPage).toBe(20);
  });

  it("Updates orderBy value", () => {
    announcementStore.setOrderBy("desc");
    expect(announcementStore.getOrderBy).toBe("desc");
  });

  it("Clears announcements array", () => {
    announcementStore.announcements = announcements;
    announcementStore.clearAnnouncements();
    expect(announcementStore.getAnnouncements).toEqual([]);
  });

  it("Clears single announcement", () => {
    announcementStore.announcement = announcement;
    announcementStore.clearAnnouncement();
    expect(announcementStore.getAnnouncement).toEqual({});
  });
});
