<template>
  <v-list-item
    v-bind="$attrs"
    @click="open"
    :disabled="!hasAuthorization"
    data-test="open-tags-btn"
  >
    <div class="d-flex align-center">
      <div class="mr-2">
        <v-icon> mdi-tag </v-icon>
      </div>

      <v-list-item-title data-test="has-tags-verification">
        {{ hasTags ? "Edit tags" : "Add Tags" }}
      </v-list-item-title>
    </div>
  </v-list-item>

  <BaseDialog v-model="showDialog">
    <v-card class="bg-v-theme-surface">
      <v-card-title class="text-h5 pa-4 bg-primary" data-test="title">
        {{ hasTags ? "Edit tags" : "Add Tags" }}
      </v-card-title>
      <v-divider />

      <v-card-text class="mt-5 w-100">
        <v-autocomplete
          v-model="selectedTags"
          v-model:menu="acMenuOpen"
          :menu-props="{ contentClass: menuContentClass, maxHeight: 320 }"
          :items="tags"
          item-title="name"
          item-value="name"
          label="Tag"
          multiple
          variant="outlined"
          data-test="deviceTag-autocomplete"
          @update:search="onSearch"
        >
          <template #item="{ item, props }">
            <v-list-item
              v-bind="{ ...props, title: undefined }"
              :key="item.value"
              @click.stop.prevent="updateTags(item.value)"
              :active="selectedTags.includes(item.value)"
              active-color="primary"
              data-test="tag-item"
            >
              <template #prepend>
                <v-checkbox :model-value="selectedTags.includes(item.value)" color="primary" hide-details />
              </template>
              <template #title>
                <v-chip>{{ item.value }}</v-chip>
              </template>
            </v-list-item>

            <v-divider />
          </template>

          <template v-slot:prepend-item>
            <div class="d-flex justify-center">
              <v-btn
                v-if="validNewTag"
                @click="createTag"
                color="primary"
                variant="text"
                data-test="create-new-tag-btn"
              >
                Create New Tag
              </v-btn>
            </div>
          </template>

          <template v-slot:selection="{ item }">
            <v-chip
              :key="item.value"
              closable
              @click:close="removeTag(item.value)"
              data-test="selected-tags"
            >
              {{ item.value }}
            </v-chip>
          </template>

          <template v-slot:append-item>
            <div ref="sentinel" data-test="tags-sentinel" style="height: 1px;" />
          </template>
        </v-autocomplete>
      </v-card-text>

      <v-card-actions>
        <v-spacer />
        <v-btn
          variant="text"
          data-test="close-btn"
          @click="close()"
          class="mr-2"
        >
          Close
        </v-btn>
      </v-card-actions>
    </v-card>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, ref, watch, nextTick, onMounted, onUnmounted } from "vue";
import axios, { AxiosError } from "axios";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import BaseDialog from "../BaseDialog.vue";
import useTagsStore from "@/store/modules/tags";
import type { Tags as StoreTags } from "@/interfaces/ITags";

type LocalTag = { name: string };

const props = defineProps<{
  deviceUid: string;
  tagsList: { name: string }[];
  hasAuthorization?: boolean;
}>();

const emit = defineEmits(["update"]);

const snackbar = useSnackbar();
const tagsStore = useTagsStore();
const tenant = computed(() => localStorage.getItem("tenant"));

const showDialog = ref(false);

const acMenuOpen = ref(false);
const menuContentClass = computed(() => `tags-ac-content-${props.deviceUid}`);

const isLoading = ref(false);
const filter = ref("");
const tagsError = ref("");

const selectedTags = ref<string[]>(
  props.tagsList.map((t) => t?.name).filter(Boolean),
);

const fetchedTags = ref<LocalTag[]>([]);
const tags = computed(() => fetchedTags.value);

const page = ref(1);
const perPage = ref(10);

const hasMore = computed(() => tagsStore.numberTags > fetchedTags.value.length);

const hasTags = computed(() => selectedTags.value.length > 0);

const sentinel = ref<HTMLElement | null>(null);
let observer: IntersectionObserver | null = null;

const cleanupObserver = () => {
  if (observer) {
    observer.disconnect();
    observer = null;
  }
};

const close = () => {
  selectedTags.value = [];
  fetchedTags.value = [];
  showDialog.value = false;
  acMenuOpen.value = false;
  cleanupObserver();
};

const encodeFilter = (search: string) => {
  if (!search) return "";
  const filterToEncodeBase64 = [
    {
      type: "property",
      params: { name: "name", operator: "contains", value: search },
    },
  ];
  return btoa(JSON.stringify(filterToEncodeBase64));
};

watch(selectedTags, (newValue) => {
  if (selectedTags.value === newValue) {
    emit("update");
  }
});

const validNewTag = computed(() => filter.value.length >= 3
  && filter.value.length <= 255
  && !tags.value.some((t) => t.name === filter.value)
  && !selectedTags.value.includes(filter.value));

const resetPagination = (): void => {
  page.value = 1;
  perPage.value = 10;
  fetchedTags.value = [];
};

const normalizeStoreItems = (arr: StoreTags[]): LocalTag[] => (arr ?? [])
  .map((t) => {
    const name = typeof t === "string" ? t : t?.name;
    return name ? ({ name } as LocalTag) : null;
  })
  .filter((t: LocalTag | null): t is LocalTag => !!t);

const loadTags = async () => {
  if (isLoading.value) return;

  isLoading.value = true;
  try {
    const encodedFilter = encodeFilter(filter.value);

    await tagsStore.autocomplete({
      tenant: tenant.value || "",
      filter: encodedFilter,
      page: page.value,
      perPage: perPage.value,
    });

    const newTags = normalizeStoreItems(tagsStore.list as unknown as StoreTags[]);
    fetchedTags.value = newTags;
  } catch (error) {
    if (axios.isAxiosError(error)) {
      const axiosError = error as AxiosError;
      switch (axiosError.response?.status) {
        case 400:
          tagsError.value = "The format is invalid. Min 3, Max 255 characters!";
          break;
        case 403:
          snackbar.showError("You are not authorized to update this tag.");
          break;
        case 406:
          tagsError.value = "The maximum capacity has reached.";
          break;
        default:
          snackbar.showError("Failed to update tags.");
          handleError(axiosError);
      }
    } else {
      snackbar.showError("Failed to update tags.");
      handleError(error);
    }
  } finally {
    isLoading.value = false;
  }
};

const open = async () => {
  showDialog.value = true;
  resetPagination();
  selectedTags.value = props.tagsList.map((t) => t?.name).filter(Boolean);
  await loadTags();
};

const onSearch = async (search: string) => {
  filter.value = search;
  resetPagination();
  await loadTags();
};

const updateTags = async (newTag: string) => {
  const addedTag = !selectedTags.value.includes(newTag);

  if (addedTag) {
    await tagsStore.pushTagToDevice({
      tenant: tenant.value || "",
      uid: props.deviceUid,
      name: newTag,
    });
    selectedTags.value.push(newTag);
  } else {
    await tagsStore.removeTagFromDevice({
      tenant: tenant.value || "",
      uid: props.deviceUid,
      name: newTag,
    });
    const index = selectedTags.value.indexOf(newTag);
    selectedTags.value.splice(index, 1);
  }
  emit("update");
};

const createTag = async () => {
  if (!validNewTag.value) return;

  try {
    await tagsStore.createTag({
      tenant: tenant.value || "",
      name: filter.value,
    });

    await tagsStore.pushTagToDevice({
      tenant: tenant.value || "",
      uid: props.deviceUid,
      name: filter.value,
    });

    selectedTags.value.push(filter.value);

    const name = filter.value;
    fetchedTags.value = [{ name }, ...fetchedTags.value.filter((t) => t.name !== name)];

    filter.value = "";
    tagsError.value = "";
  } catch (error) {
    console.error("Error creating tag:", error);
  }
};

const removeTag = async (tag: string) => {
  selectedTags.value = selectedTags.value.filter((t) => t !== tag);

  await tagsStore.removeTagFromDevice({
    tenant: tenant.value || "",
    uid: props.deviceUid,
    name: tag,
  });
};

const bumpPerPageAndLoad = async () => {
  if (!hasMore.value || isLoading.value) return;
  perPage.value += 10;
  await loadTags();
};

const getMenuRootEl = (): HTMLElement | null => document.querySelector(`.${menuContentClass.value}`) as HTMLElement | null;

const setupObserver = () => {
  cleanupObserver();

  const root = getMenuRootEl();
  if (!root || !sentinel.value) return;

  observer = new IntersectionObserver(
    (entries) => {
      const entry = entries[0];
      if (entry?.isIntersecting) bumpPerPageAndLoad();
    },
    { root, threshold: 1.0 },
  );

  observer.observe(sentinel.value);
};

watch(acMenuOpen, async (open) => {
  if (open) {
    await nextTick();
    setupObserver();
  } else {
    cleanupObserver();
  }
});

watch(showDialog, (open) => {
  if (!open) {
    acMenuOpen.value = false;
    cleanupObserver();
  }
});

onMounted(async () => {
  await loadTags();
});

onUnmounted(() => {
  cleanupObserver();
});

defineExpose({ updateTags, loadTags, createTag, removeTag, selectedTags, fetchedTags });
</script>
