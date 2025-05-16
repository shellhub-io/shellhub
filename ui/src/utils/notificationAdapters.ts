import { IDevice } from "@/interfaces/IDevice";
import { IContainer } from "@/interfaces/IContainer";
import { INotification, NotificationType } from "@/interfaces/INotification";

export const deviceToNotification = (device: IDevice): INotification => ({
  id: device.uid,
  type: NotificationType.DEVICE,
  data: device,
});

export const containerToNotification = (container: IContainer): INotification => ({
  id: container.uid,
  type: NotificationType.CONTAINER,
  data: container,
});
