// stores/tags.ts
import { defineStore } from "pinia";
import { ref, computed } from "vue";
import * as apiTags from "../api/tags";
import { Tags } from "@/interfaces/ITags";
import { UpdateTagRequest } from "@/api/client";

const useTagsStore = defineStore("tags", () => {
  const tags = ref<Array<Tags>>([]);
  const numberTags = ref(0);
  const page = ref(1);
  const perPage = ref(10);
  const filter = ref<string | undefined>("");
  const selected = ref<{
    device: Array<Tags>;
    container: Array<Tags>;
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

  const setTags = (res) => {
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
    tag: Tags;
  }) => {
    const list = selected.value[variant];
    const exists = list.includes(tag);
    if (exists) {
      selected.value[variant] = list.filter((t) => t !== tag);
    } else {
      selected.value[variant] = [...list, tag];
    }
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
      const res = await apiTags.getTags(tenant, filter, page, perPage);
      setTags(res);
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
      const res = await apiTags.getTags(tenant, filter, page.value, perPage.value);
      setTags(res);
      setFilter(filter);
    } catch (error) {
      clearListTags();
      throw error;
    }
  };

  const autocomplete = async ({
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
      const res = await apiTags.getTags(tenant, filter, page, perPage);
      setTags(res);
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
    try {
      await apiTags.createTag(tenant, name);
    } catch (error) {
      console.error(error);
      throw error;
    }
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
    try {
      await apiTags.updateTag(tenant, currentName, newName);
    } catch (error) {
      console.error(error);
      throw error;
    }
  };

  const removeTag = async ({
    tenant,
    currentName,
  }: {
    tenant: string;
    currentName: string;
  }) => {
    try {
      await apiTags.removeTag(tenant, currentName);
    } catch (error) {
      console.error(error);
      throw error;
    }
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
    await apiTags.pushTagToDevice(tenant, uid, name);
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
    await apiTags.removeTagFromDevice(tenant, uid, name);
  };

  const pushTagToFirewallRule = async ({
    tenant,
    name,
  }: {
    tenant: string;
    name: string;
  }) => {
    await apiTags.pushTagToFirewallRule(tenant, name);
  };

  const removeTagFromFirewallRule = async ({
    tenant,
    name,
  }: {
    tenant: string;
    name: string;
  }) => {
    await apiTags.removeTagFromFirewallRule(tenant, name);
  };

  const pushTagToPublicKey = async ({
    tenant,
    fingerprint,
    name,
  }: {
    tenant: string;
    fingerprint: string;
    name: string;
  }) => {
    await apiTags.pushTagToPublicKey(tenant, fingerprint, name);
  };

  const removeTagFromPublicKey = async ({
    tenant,
    fingerprint,
    name,
  }: {
    tenant: string;
    fingerprint: string;
    name: string;
  }) => {
    await apiTags.removeTagFromPublicKey(tenant, fingerprint, name);
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
    pushTagToFirewallRule,
    removeTagFromFirewallRule,
    pushTagToPublicKey,
    removeTagFromPublicKey,
  };
});

export default useTagsStore;
