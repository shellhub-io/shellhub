import { IDevice } from "@/interfaces/IDevice";

export enum NotificationType {
  DEVICE = "device",
  CONTAINER = "container"
}

type NotificationData = IDevice;

export interface INotification {
  id: string;
  type: NotificationType;
  data: NotificationData;
}
