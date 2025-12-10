import { tagsApi } from "@/api/http";
import { IUpdateTagName } from "@/interfaces/ITags";

export const createTag = async (name: string) => tagsApi.createTag({ name });

export const updateTag = async (
  currentName: string,
  newName: IUpdateTagName,
) => tagsApi.updateTag(currentName, newName);

export const removeTag = async (name: string) => tagsApi.deleteTag(name);

export const removeTagFromDevice = async (uid: string, name: string) => tagsApi.pullTagFromDevice(uid, name);

export const pushTagToDevice = async (uid: string, name: string) => tagsApi.pushTagToDevice(uid, name);

export const getTags = async (
  page: number,
  perPage: number,
  filter?: string,
) => tagsApi.getTags(filter, page, perPage);
