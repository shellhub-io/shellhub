import { UpdateTagRequest } from "@/api/client";
import { tagsApi } from "@/api/http";

export const createTag = async (name: string) => tagsApi.createTag({ name });

export const updateTag = async (
  currentName: string,
  newName: UpdateTagRequest,
) => tagsApi.updateTag(currentName, newName);

export const removeTag = async (name: string) => tagsApi.deleteTag(name);

export const removeTagFromDevice = async (uid: string, name: string) => tagsApi.pullTagFromDevice(uid, name);

export const pushTagToDevice = async (uid: string, name: string) => tagsApi.pushTagToDevice(uid, name);

export const getTags = async (
  filter: string | undefined,
  page: number,
  perPage: number,
) => tagsApi.getTags(filter, page, perPage);
