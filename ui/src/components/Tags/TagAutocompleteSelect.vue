<template>
  <v-autocomplete
    v-model="selectedTags"
    v-model:menu="isAutocompleteMenuOpen"
    :menu-props="{ contentClass: menuContentClass, maxHeight: 320 }"
    :items="tags"
    variant="outlined"
    item-title="name"
    item-value="name"
    attach
    chips
    class="mb-4"
    hide-details="auto"
    label="Tags"
    density="comfortable"
    :error-messages="tagSelectorErrorMessage"
    multiple
    data-test="tags-selector"
    @update:model-value="validateSelectedTags"
    @update:search="handleSearch"
  >
    <template #append-item>
      <div
        ref="sentinel"
        data-test="tags-sentinel"
        style="height: 1px"
      />
    </template>
  </v-autocomplete>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, nextTick, onUnmounted } from "vue";
import { useIntersectionObserver } from "@vueuse/core";
import { ITag } from "@/interfaces/ITags";
import useSnackbar from "@/helpers/snackbar";
import useTagsStore from "@/store/modules/tags";
import handleError from "@/utils/handleError";

const tagsStore = useTagsStore();
const snackbar = useSnackbar();

const tags = ref<ITag[]>([]);
const hasMoreTagsToLoad = computed(() => tagsStore.numberTags > tags.value.length);
const selectedTags = defineModel<string[]>("selectedTags", { required: true });
const tagSelectorErrorMessage = defineModel<string>("tagSelectorErrorMessage", { required: true });
const isAutocompleteMenuOpen = ref(false);

const sentinel = ref<HTMLElement | null>(null);
const menuContentClass = "tags-autocomplete-content";

const itemsPerPage = ref(10);
const filter = ref("");
const isLoading = ref(false);

const validateSelectedTags = async () => {
  await nextTick(); // Ensure DOM and v-model updates are complete before validation
  const list = selectedTags.value;
  if (list.length > 3) tagSelectorErrorMessage.value = "You can select up to three tags only";
  else if (list.length === 0) tagSelectorErrorMessage.value = "You must choose at least one tag";
  else if (list.length <= 3) tagSelectorErrorMessage.value = "";
};

const encodeFilter = (filterQuery: string) => {
  if (!filterQuery) return "";
  const filterQueryForEncoding = [{ type: "property", params: { name: "name", operator: "contains", value: filterQuery } }];
  return Buffer.from(JSON.stringify(filterQueryForEncoding)).toString("base64");
};

const resetPagination = () => {
  itemsPerPage.value = 10;
};

const loadTags = async () => {
  if (isLoading.value) return;
  isLoading.value = true;

  try {
    await tagsStore.autocomplete({
      tenant: localStorage.getItem("tenant") || "",
      filter: encodeFilter(filter.value),
      perPage: itemsPerPage.value,
    });
    tags.value = tagsStore.list;
  } catch (error) {
    snackbar.showError("Failed to load tags.");
    handleError(error);
  } finally {
    isLoading.value = false;
  }
};

const handleSearch = async (filterQuery = "") => {
  filter.value = filterQuery;
  resetPagination();
  await loadTags();
};

const bumpPerPageAndLoad = async () => {
  if (!hasMoreTagsToLoad.value || isLoading.value) return;
  itemsPerPage.value += 10;
  await loadTags();
};

const getMenuRootEl = (): HTMLElement | null => document.querySelector(`.${menuContentClass}`);

useIntersectionObserver(
  sentinel,
  ([{ isIntersecting }]) => { if (isIntersecting) void bumpPerPageAndLoad(); },
  { root: getMenuRootEl, threshold: 1.0 },
);

onMounted(async () => {
  resetPagination();
  await loadTags();
  await validateSelectedTags();
});

onUnmounted(() => {
  resetPagination();
  isAutocompleteMenuOpen.value = false;
});
</script>
