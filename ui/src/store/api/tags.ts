import { tagsApi } from "../../api/http";

interface UpdateTagResponse {
  oldTag: string;
  newTag: string;
}

export const updateTag = async (data : UpdateTagResponse) => tagsApi.renameTag(data.oldTag, { tag: data.newTag });

export const removeTag = async (tag : string) => tagsApi.deleteTag(tag);

export const getTags = async () => tagsApi.getTags();
