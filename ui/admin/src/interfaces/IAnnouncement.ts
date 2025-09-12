import { IAnnouncement, IAnnouncementShort } from "@/interfaces/IAnnouncement";

export type IAdminAnnouncementShort = IAnnouncementShort;

export type IAdminAnnouncement = IAnnouncement;

export type IAdminAnnouncementRequestBody = Pick<IAnnouncement, "title" | "content">;
