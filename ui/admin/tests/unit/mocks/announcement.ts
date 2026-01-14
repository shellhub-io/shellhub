import { IAdminAnnouncement, IAdminAnnouncementShort } from "@admin/interfaces/IAnnouncement";

export const mockAnnouncement: IAdminAnnouncement = {
  uuid: "announcement-123",
  title: "Test Announcement",
  content: "## ShellHub new features \n - New feature 1 \n - New feature 2",
  date: "2026-01-14T08:00:00.000Z",
};

export const mockAnnouncementShort: IAdminAnnouncementShort = {
  uuid: "announcement-123",
  title: "Test Announcement",
  date: "2026-01-14T08:00:00.000Z",
};

export const mockAnnouncements: IAdminAnnouncementShort[] = [
  { ...mockAnnouncementShort, uuid: "announcement-1", title: "Announcement One" },
  { ...mockAnnouncementShort, uuid: "announcement-2", title: "Announcement Two" },
  { ...mockAnnouncementShort, uuid: "announcement-3", title: "Announcement Three" },
];
