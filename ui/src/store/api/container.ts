import { IContainerPostTag, IContainerRename, IUpdateContainerTag, IUpdateContainerTags } from "@/interfaces/IContainer";
import { containersApi, tagsApi } from "../../api/http";

export const postTag = async (data: IContainerPostTag) => tagsApi.createContainerTag(data.uid, data.name);

export const fetchContainers = async (
  page : number,
  perPage: number,
  filter : string | undefined,
  status : "accepted" | "rejected" | "pending" | "unused",
  sortStatusField : string | undefined,
  sortStatusString : "asc" | "desc" | "",
) => {
  if (sortStatusField && sortStatusString) {
    return containersApi.getContainers(
      filter,
      page,
      perPage,
      status,
      sortStatusField,
      sortStatusString,
    );
  }

  return containersApi.getContainers(filter, page, perPage, status);
};

export const getContainer = async (uid : string) => containersApi.getContainer(uid);

export const renameContainer = async (data : IContainerRename) => containersApi.updateContainer(data.uid, data.name);

export const acceptContainer = async (uid : string) => containersApi.updateContainerStatus(uid, "accept");

export const rejectContainer = async (uid : string) => containersApi.updateContainerStatus(uid, "reject");

export const updateContainerTag = async (data : IUpdateContainerTags) => containersApi.updateTagsContainer(data.uid, data.tags);

export const deleteContainerTag = async (data : IUpdateContainerTag) => containersApi.deleteContainerTag(data.uid, data.tags);

export const removeContainer = async (uid : string) => containersApi.deleteContainer(uid);
