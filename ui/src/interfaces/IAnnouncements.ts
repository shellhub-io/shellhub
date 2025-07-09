export interface IAnnouncementShort {
  uuid: string;
  title: string;
  date: string;
}

export interface IAnnouncement extends IAnnouncementShort {
  content: string;
}
