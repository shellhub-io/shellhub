import { defineStore } from "pinia";
import { ref } from "vue";
import * as tagsApi from "../api/tags";

const useTagsStore = defineStore("tags", () => {
  const tags = ref<Array<string>>([]);
  const tagsCount = ref<number>(0);

  const fetchTags = async () => {
    const res = await tagsApi.getTags();
    tags.value = res.data;
    tagsCount.value = parseInt(res.headers["x-total-count"], 10);
  };

  const updateTag = async (data: { oldTag: string, newTag: string }) => {
    await tagsApi.updateTag(data);
  };

  const setTags = (newTags: Array<string>) => {
    tags.value = newTags;
    tagsCount.value = newTags.length;
  };

  const removeTag = async (name: string) => {
    await tagsApi.removeTag(name);
  };

  return {
    tags,
    tagsCount,
    fetchTags,

    updateTag,
    setTags,
    removeTag,
  };
});

export default useTagsStore;
