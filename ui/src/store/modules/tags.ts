import { defineStore } from "pinia";
import { ref } from "vue";
import * as tagsApi from "../api/tags";
import { ITag, IUpdateTagName } from "@/interfaces/ITags";

const useTagsStore = defineStore("tags", () => {
  const tags = ref<ITag[]>([]);
  const tagCount = ref(0);
  const showTags = ref(false);
  const selectedTags = ref<ITag[]>([]);

  const getTagName = (tag: ITag) => (typeof tag === "string" ? tag : tag.name);

  const toggleSelectedTag = (tag: ITag) => {
    const selectedTagName = getTagName(tag);
    const isTagSelected = selectedTags.value.some((tag) => getTagName(tag) === selectedTagName);

    selectedTags.value = isTagSelected
      ? selectedTags.value.filter((tag) => getTagName(tag) !== selectedTagName)
      : [...selectedTags.value, tag];
  };

  const fetchTagList = async (data?: { filter?: string; page?: number; perPage?: number }) => {
    try {
      const res = await tagsApi.getTags(
        data?.page || 1,
        data?.perPage || 10,
        data?.filter,
      );
      tags.value = res.data as ITag[];
      tagCount.value = parseInt(res.headers["x-total-count"] as string, 10) || 0;
    } catch (error) {
      tags.value = [];
      tagCount.value = 0;
      throw error;
    }
  };

  const setTagListVisibility = async () => {
    const { headers } = await tagsApi.getTags(1, 1);
    const count = parseInt(headers["x-total-count"] as string, 10) || 0;
    if (count) showTags.value = true;
  };

  const createTag = async (name: string) => {
    await tagsApi.createTag(name);
  };

  const updateTag = async (currentName: string, newName: IUpdateTagName) => {
    await tagsApi.updateTag(currentName, newName);
  };

  const deleteTag = async (name: string) => {
    await tagsApi.removeTag(name);
  };

  const addTagToDevice = async (uid: string, name: string) => {
    await tagsApi.pushTagToDevice(uid, name);
  };

  const removeTagFromDevice = async (uid: string, name: string) => {
    await tagsApi.removeTagFromDevice(uid, name);
  };

  return {
    tags,
    tagCount,
    showTags,
    selectedTags,

    toggleSelectedTag,
    fetchTagList,
    setTagListVisibility,
    createTag,
    updateTag,
    deleteTag,
    addTagToDevice,
    removeTagFromDevice,
  };
});

export default useTagsStore;
