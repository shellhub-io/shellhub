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

  <FormDialog
    v-model="showDialog"
    @close="close"
    @cancel="close"
    @confirm="close"
    :title="hasTags ? 'Edit tags' : 'Add Tags'"
    icon="mdi-tag"
    confirm-text=""
    :confirm-disabled="true"
    cancel-text="Close"
    cancel-data-test="close-btn"
    data-test="tags-form-dialog"
  >
    <div class="px-6 pt-4">
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
        @update:modelValue="onTagSelectionChanged"
      >
        <template #item="{ item, props }">
          <v-list-item
            v-bind="{ ...props, title: undefined }"
            :key="item.value"
            :active="selectedTags.includes(item.value)"
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

        <template #prepend-item>
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

        <template #selection="{ item }">
          <v-chip
            :key="item.value"
            closable
            @click:close="removeTag(item.value)"
            data-test="selected-tags"
          >
            {{ item.value }}
          </v-chip>
        </template>

        <template #append-item>
          <div ref="sentinel" data-test="tags-sentinel" style="height: 1px;" />
        </template>
      </v-autocomplete>
    </div>
  </FormDialog>
</template>

<script setup lang="ts">
import { computed, ref, watch, nextTick, onMounted, onUnmounted } from "vue";
import axios, { AxiosError } from "axios";
import FormDialog from "@/components/Dialogs/FormDialog.vue";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import useTagsStore from "@/store/modules/tags";
import type { ITag as StoreTags } from "@/interfaces/ITags";

type LocalTag = { name: string };

const props = defineProps<{
  deviceUid: string;
  tagsList: { name: string }[];
  hasAuthorization?: boolean;
}>();

const emit = defineEmits(["update"]);

const snackbar = useSnackbar();
const tagsStore = useTagsStore();
const tenant = computed(() => localStorage.getItem("tenant") || "");

const showDialog = ref(false);

const acMenuOpen = ref(false);
const menuContentClass = computed(() => `tags-ac-content-${props.deviceUid}`);

const isLoading = ref(false);
const filter = ref("");
const tagsError = ref("");

const selectedTags = ref<string[]>(
  props.tagsList.map((t) => t?.name).filter(Boolean),
);

const previousTags = ref<string[]>([...selectedTags.value]);

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

const validNewTag = computed(
  () => filter.value.length >= 3
    && filter.value.length <= 255
    && !tags.value.some((t) => t.name === filter.value)
    && !selectedTags.value.includes(filter.value),
);

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
      tenant: tenant.value,
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
  previousTags.value = [...selectedTags.value];
  await loadTags();
};

const onSearch = async (search: string) => {
  filter.value = search;
  resetPagination();
  await loadTags();
};

const onTagSelectionChanged = async (newTags: string[]) => {
  const oldTags = previousTags.value;

  const added = newTags.filter((tag) => !oldTags.includes(tag));
  const removed = oldTags.filter((tag) => !newTags.includes(tag));

  try {
    await Promise.all(
      added.map((tag) => tagsStore.pushTagToDevice({
        tenant: tenant.value,
        uid: props.deviceUid,
        name: tag,
      })),
    );

    await Promise.all(
      removed.map((tag) => tagsStore.removeTagFromDevice({
        tenant: tenant.value,
        uid: props.deviceUid,
        name: tag,
      })),
    );

    previousTags.value = [...newTags];
    emit("update");
  } catch (error) {
    snackbar.showError("Failed to update tags.");
    handleError(error);
  }
};

const updateTags = async (tag: string) => {
  const isSelected = selectedTags.value.includes(tag);
  const newTags = isSelected
    ? selectedTags.value.filter((t) => t !== tag)
    : [...selectedTags.value, tag];

  selectedTags.value = newTags;
  await onTagSelectionChanged(newTags);
};

const createTag = async () => {
  if (!validNewTag.value) return;

  try {
    await tagsStore.createTag({
      tenant: tenant.value,
      name: filter.value,
    });

    const name = filter.value;
    fetchedTags.value = [{ name }, ...fetchedTags.value.filter((t) => t.name !== name)];

    const newTags = Array.from(new Set([...selectedTags.value, name]));
    selectedTags.value = newTags;
    await onTagSelectionChanged(newTags);

    filter.value = "";
    tagsError.value = "";
  } catch (error) {
    snackbar.showError("Failed to create tag.");
    handleError(error);
  }
};

const removeTag = async (tag: string) => {
  const newTags = selectedTags.value.filter((t) => t !== tag);
  selectedTags.value = newTags;
  await onTagSelectionChanged(newTags);
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

const close = () => {
  showDialog.value = false;
  acMenuOpen.value = false;
  cleanupObserver();
  resetPagination();
  filter.value = "";
  tagsError.value = "";
};

watch(acMenuOpen, async (openVal) => {
  if (openVal) {
    await nextTick();
    setupObserver();
  } else {
    cleanupObserver();
  }
});

watch(showDialog, (openVal) => {
  if (!openVal) {
    acMenuOpen.value = false;
    cleanupObserver();
  }
});

onMounted(async () => {
  await loadTags();
  previousTags.value = [...selectedTags.value];
});

onUnmounted(() => {
  cleanupObserver();
});

defineExpose({ loadTags, createTag, removeTag, updateTags, selectedTags, tags });
</script>
