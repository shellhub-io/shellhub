import { describe, expect, it } from "vitest";
import { setActivePinia, createPinia } from "pinia";
import useAnnouncementStore from "@admin/store/modules/announcement";

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

const [firstAnnouncement] = announcements;
const announcementCount = 2;

describe("Announcement Store (Pinia)", () => {
  setActivePinia(createPinia());
  const announcementStore = useAnnouncementStore();

  it("Returns default state", () => {
    expect(announcementStore.announcements).toEqual([]);
    expect(announcementStore.announcement).toEqual({});
    expect(announcementStore.announcementCount).toBe(0);
  });

  it("Sets announcements and updates count", () => {
    announcementStore.announcements = announcements;
    announcementStore.announcementCount = announcementCount;
    expect(announcementStore.announcements).toEqual(announcements);
    expect(announcementStore.announcementCount).toBe(announcementCount);
  });

  it("Sets a single announcement", () => {
    announcementStore.announcement = firstAnnouncement;
    expect(announcementStore.announcement).toEqual(firstAnnouncement);
  });
});
