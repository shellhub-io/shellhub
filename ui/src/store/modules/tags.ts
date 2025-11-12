import { defineStore } from "pinia";
import { ref, computed } from "vue";
import * as tagsApi from "../api/tags";
import { ITag } from "@/interfaces/ITags";
import { UpdateTagRequest } from "@/api/client";

// Temporary solution until module refactoring
interface TagsResponse {
  data: Array<ITag>;
  headers: Record<string, string>;
}

const useTagsStore = defineStore("tags", () => {
  const tags = ref<Array<ITag>>([]);
  const numberTags = ref(0);
  const page = ref(1);
  const perPage = ref(10);
  const filter = ref<string | undefined>("");
  const selected = ref<{
    device: Array<ITag>;
    container: Array<ITag>;
  }>({
    device: [],
    container: [],
  });

  const list = computed(() => tags.value);
  const getNumberTags = computed(() => numberTags.value);
  const getPage = computed(() => page.value);
  const getPerPage = computed(() => perPage.value);
  const getFilter = computed(() => filter.value);

  const getSelected = (variant: "device" | "container") => selected.value[variant];

  const setTags = (res: TagsResponse) => {
    tags.value = res.data;
    numberTags.value = parseInt(res.headers["x-total-count"], 10);
  };

  const setPagePerPage = (data: { page: number; perPage: number }) => {
    page.value = data.page;
    perPage.value = data.perPage;
  };

  const setFilter = (value: string) => {
    filter.value = value;
  };

  const clearListTags = () => {
    tags.value = [];
    numberTags.value = 0;
  };

  const setSelected = ({
    variant,
    tag,
  }: {
    variant: "device" | "container";
    tag: ITag;
  }) => {
    const toName = (tag: ITag) => (typeof tag === "string" ? tag : tag.name);
    const list = selected.value[variant];
    const name = toName(tag);
    const exists = list.some((t) => toName(t) === name);

    selected.value[variant] = exists
      ? list.filter((t) => toName(t) !== name)
      : [...list, tag];
  };

  const clearSelected = (variant: "device" | "container") => {
    selected.value[variant] = [];
  };

  const fetch = async ({
    tenant,
    filter,
    page,
    perPage,
  }: {
    tenant: string;
    filter: string;
    page: number;
    perPage: number;
  }) => {
    try {
      const res = await tagsApi.getTags(tenant, filter, page, perPage);
      setTags(res as unknown as TagsResponse);
      setPagePerPage({ page, perPage });
      setFilter(filter);
    } catch (error) {
      clearListTags();
      throw error;
    }
  };

  const search = async ({
    tenant,
    filter,
  }: {
    tenant: string;
    filter: string;
  }) => {
    try {
      const res = await tagsApi.getTags(tenant, filter, page.value, perPage.value);
      setTags(res as unknown as TagsResponse);
      setFilter(filter);
    } catch (error) {
      clearListTags();
      throw error;
    }
  };

  const autocomplete = async ({
    tenant,
    filter,
    perPage,
  }: {
    tenant: string;
    filter: string;
    perPage: number;
  }) => {
    try {
      const res = await tagsApi.getTags(tenant, filter, 1, perPage);
      setTags(res as unknown as TagsResponse);
      setFilter(filter);
    } catch (error) {
      clearListTags();
      throw error;
    }
  };

  const createTag = async ({
    tenant,
    name,
  }: {
    tenant: string;
    name: string;
  }) => {
    await tagsApi.createTag(tenant, name);
  };

  const editTag = async ({
    tenant,
    currentName,
    newName,
  }: {
    tenant: string;
    currentName: string;
    newName: UpdateTagRequest;
  }) => {
    await tagsApi.updateTag(tenant, currentName, newName);
  };

  const removeTag = async ({
    tenant,
    currentName,
  }: {
    tenant: string;
    currentName: string;
  }) => {
    await tagsApi.removeTag(tenant, currentName);
  };

  const pushTagToDevice = async ({
    tenant,
    uid,
    name,
  }: {
    tenant: string;
    uid: string;
    name: string;
  }) => {
    await tagsApi.pushTagToDevice(tenant, uid, name);
  };

  const removeTagFromDevice = async ({
    tenant,
    uid,
    name,
  }: {
    tenant: string;
    uid: string;
    name: string;
  }) => {
    await tagsApi.removeTagFromDevice(tenant, uid, name);
  };

  return {
    // State
    tags,
    numberTags,
    page,
    perPage,
    filter,
    selected,

    // Getters
    list,
    getNumberTags,
    getPage,
    getPerPage,
    getFilter,
    getSelected,

    // Mutations/Actions
    setSelected,
    clearSelected,
    setFilter,
    clearListTags,

    // Async API Actions
    fetch,
    search,
    autocomplete,
    createTag,
    editTag,
    removeTag,
    pushTagToDevice,
    removeTagFromDevice,
  };
});

export default useTagsStore;
