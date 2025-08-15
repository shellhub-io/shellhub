import { IContainerPostTag, IContainerRename, IUpdateContainerTags } from "@/interfaces/IContainer";
import { containersApi, tagsApi } from "@/api/http";

export const postTag = async (data: IContainerPostTag) => tagsApi.createContainerTag(data.uid, data.name);

export const fetchContainers = async (
  page: number,
  perPage: number,
  status: "accepted" | "rejected" | "pending",
  filter?: string,
  sortField?: string,
  sortOrder?: "asc" | "desc",
) => containersApi.getContainers(
  filter,
  page,
  perPage,
  status,
  sortField,
  sortOrder,
);

export const getContainer = async (uid: string) => containersApi.getContainer(uid);

export const renameContainer = async (data: IContainerRename) => containersApi.updateContainer(data.uid, data.name);

export const acceptContainer = async (uid: string) => containersApi.updateContainerStatus(uid, "accept");

export const rejectContainer = async (uid: string) => containersApi.updateContainerStatus(uid, "reject");

export const updateContainerTag = async (data: IUpdateContainerTags) => containersApi.updateTagsContainer(data.uid, data.tags);

export const removeContainer = async (uid: string) => containersApi.deleteContainer(uid);
