import { IContainer } from "@/interfaces/IContainer";
import { IDevice } from "@/interfaces/IDevice";

export enum NotificationType {
  DEVICE = "device",
  CONTAINER = "container"
}

type NotificationData = IDevice | IContainer;

export interface INotification {
  id: string;
  type: NotificationType;
  data: NotificationData;
}
