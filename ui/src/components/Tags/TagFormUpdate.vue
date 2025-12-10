<template>
  <v-list-item
    v-bind="$attrs"
    :disabled="!hasAuthorization"
    data-test="open-tags-btn"
    @click="open"
  >
    <div class="d-flex align-center ga-2">
      <v-icon icon="mdi-tag" />
      <v-list-item-title data-test="has-tags-verification">
        {{ hasTags ? "Edit tags" : "Add Tags" }}
      </v-list-item-title>
    </div>
  </v-list-item>

  <FormDialog
    v-model="showDialog"
    :title="hasTags ? 'Edit Tags' : 'Add Tags'"
    icon="mdi-tag"
    confirm-text="Save"
    :confirm-disabled="!hasChanges"
    cancel-text="Cancel"
    cancel-data-test="cancel-btn"
    data-test="tags-form-dialog"
    @close="close"
    @cancel="close"
    @confirm="saveTags"
  >
    <div class="pa-6">
      <v-autocomplete
        v-model="selectedTags"
        v-model:menu="isAutocompleteMenuOpen"
        v-model:search="searchedTag"
        :menu-props="{ contentClass: menuContentClass, maxHeight: 320 }"
        :items="availableTags"
        item-title="name"
        item-value="name"
        label="Tags"
        multiple
        variant="outlined"
        hide-details="auto"
        :messages="selectedTags.length >= 3 ? 'Maximum of 3 tags reached' : ''"
        data-test="device-tags-autocomplete"
        @update:search="handleSearch"
        @update:model-value="handleTagSelectionChange"
      >
        <template #item="{ item, props }">
          <v-list-item
            v-bind="{ ...props, title: undefined }"
            :key="item.value"
            :active="selectedTags.includes(item.value)"
            :disabled="selectedTags.length >= 3 && !selectedTags.includes(item.value)"
            data-test="tag-item"
          >
            <template #prepend>
              <v-checkbox
                :model-value="selectedTags.includes(item.value)"
                color="primary"
                hide-details
              />
            </template>
            <template #title>
              <v-chip :text="item.value" />
            </template>
          </v-list-item>

          <v-divider />
        </template>

        <template #prepend-item>
          <div class="d-flex justify-center">
            <v-btn
              v-if="isNewTagValid && selectedTags.length < 3"
              color="primary"
              variant="text"
              data-test="create-new-tag-btn"
              text="Create New Tag"
              @click="createTag"
            />
          </div>
        </template>

        <template #selection="{ item }">
          <v-chip
            :key="item.value"
            closable
            data-test="selected-tags"
            :text="item.value"
            @click:close="removeTag(item.value)"
          />
        </template>

        <template #append-item>
          <div
            ref="sentinel"
            data-test="tags-sentinel"
            style="height: 1px"
          />
        </template>
      </v-autocomplete>
    </div>
  </FormDialog>
</template>

<script setup lang="ts">
import { computed, ref, onMounted } from "vue";
import { useIntersectionObserver } from "@vueuse/core";
import FormDialog from "@/components/Dialogs/FormDialog.vue";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import useTagsStore from "@/store/modules/tags";
import { ITag } from "@/interfaces/ITags";

const props = defineProps<{
  deviceUid: string;
  tagsList: { name: string }[];
  hasAuthorization?: boolean;
}>();

const emit = defineEmits(["update"]);

const snackbar = useSnackbar();
const tagsStore = useTagsStore();

const showDialog = ref(false);
const isAutocompleteMenuOpen = ref(false);
const menuContentClass = "tags-ac-content";

const isLoading = ref(false);
const searchedTag = ref("");
const tagsError = ref("");

const selectedTags = ref<string[]>(props.tagsList.map((t) => t?.name).filter(Boolean));
const initialTags = ref<string[]>([...selectedTags.value]);
const availableTags = computed({
  get: () => tagsStore.tags,
  set: (value) => { tagsStore.tags = value; },
});
const perPage = ref(10);
const hasMoreTagsToLoad = computed(() => tagsStore.tagCount > availableTags.value.length);
const hasTags = computed(() => selectedTags.value.length > 0);
const hasChanges = computed(() => {
  if (selectedTags.value.length !== initialTags.value.length) return true;
  return !selectedTags.value.every((tag) => initialTags.value.includes(tag));
});
const sentinel = ref<HTMLElement | null>(null);
const isNewTagValid = computed(
  () =>
    searchedTag.value.length >= 3
    && searchedTag.value.length <= 255
    && !availableTags.value.some((t) => t.name === searchedTag.value)
    && !selectedTags.value.includes(searchedTag.value),
);

const encodeFilter = () => {
  if (!searchedTag.value) return "";
  const filterObject = [{
    type: "property",
    params: { name: "name", operator: "contains", value: searchedTag.value },
  }];
  return Buffer.from(JSON.stringify(filterObject)).toString("base64");
};

const resetPagination = () => { perPage.value = 10; };

const loadTags = async () => {
  if (isLoading.value) return;

  isLoading.value = true;
  try {
    await tagsStore.fetchTagList({
      filter: encodeFilter(),
      perPage: perPage.value,
    });
  } catch (error) {
    snackbar.showError("Failed to load tags.");
    handleError(error);
  } finally {
    isLoading.value = false;
  }
};

const open = async () => {
  showDialog.value = true;
  resetPagination();
  selectedTags.value = props.tagsList.map((t) => t?.name).filter(Boolean);
  initialTags.value = selectedTags.value;
  await loadTags();
};

const handleSearch = async () => {
  resetPagination();
  await loadTags();
};

const removeTag = (tagToRemove: string) => {
  selectedTags.value = selectedTags.value.filter((tag) => tag !== tagToRemove);
};

const handleTagSelectionChange = (newTags: string[]) => {
  if (newTags.length > 3) {
    selectedTags.value = selectedTags.value.slice(0, 3);
    snackbar.showError("Maximum of 3 tags allowed.");
    return;
  }
};

const createTag = async () => {
  if (!isNewTagValid.value) return;

  if (selectedTags.value.length >= 3) {
    snackbar.showError("Maximum of 3 tags allowed.");
    return;
  }

  try {
    await tagsStore.createTag(searchedTag.value);

    const newTagName = searchedTag.value;
    availableTags.value.push({ name: newTagName } as ITag);
    selectedTags.value.push(newTagName);
    searchedTag.value = "";
    tagsError.value = "";
  } catch (error) {
    snackbar.showError("Failed to create tag.");
    handleError(error);
  }
};

const saveTags = async () => {
  try {
    const tagsToAdd = selectedTags.value.filter((tag) => !initialTags.value.includes(tag));
    const tagsToRemove = initialTags.value.filter((tag) => !selectedTags.value.includes(tag));
    await Promise.all([
      ...tagsToAdd.map((tag) => tagsStore.addTagToDevice(props.deviceUid, tag)),
      ...tagsToRemove.map((tag) => tagsStore.removeTagFromDevice(props.deviceUid, tag)),
    ]);

    initialTags.value = selectedTags.value;
    snackbar.showSuccess("Tags updated successfully.");
    emit("update");
    close();
  } catch (error) {
    snackbar.showError("Failed to update tags.");
    handleError(error);
  }
};

const bumpPerPageAndLoad = async () => {
  if (!hasMoreTagsToLoad.value || isLoading.value) return;
  perPage.value += 10;
  await loadTags();
};

const getMenuRootEl = (): HTMLElement | null => document.querySelector(`.${menuContentClass}`);

useIntersectionObserver(
  sentinel,
  ([{ isIntersecting }]) => { if (isIntersecting) void bumpPerPageAndLoad(); },
  { root: getMenuRootEl, threshold: 1.0 },
);

const close = () => {
  showDialog.value = false;
  isAutocompleteMenuOpen.value = false;
  resetPagination();
  searchedTag.value = "";
  tagsError.value = "";
};

onMounted(async () => {
  await loadTags();
  initialTags.value = selectedTags.value;
});

defineExpose({
  loadTags,
  createTag,
  saveTags,
  removeTag,
  selectedTags,
  availableTags,
});
</script>
