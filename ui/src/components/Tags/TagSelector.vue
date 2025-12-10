<template>
  <div class="mr-4">
    <v-menu
      v-model="isMenuOpen"
      location="bottom"
      v-bind="$attrs"
      scrim
      eager
    >
      <template #activator="{ props }">
        <v-badge
          bordered
          color="primary"
          :content="selectedTags.length"
          :value="selectedTags.length"
        >
          <v-btn
            v-bind="props"
            data-test="tags-btn"
            color="primary"
            variant="outlined"
            append-icon="mdi-chevron-down"
            text="Tags"
          />
        </v-badge>
      </template>

      <div class="bg-v-theme-surface">
        <v-list
          v-if="fetchedTags.length > 0"
          ref="scrollArea"
          density="compact"
          style="max-height: 320px; overflow-y: auto"
        >
          <div
            v-for="tag in fetchedTags"
            :key="tag.name"
          >
            <v-list-item
              data-test="tag-item"
              @click="selectTag(tag)"
            >
              <v-list-item-action>
                <v-checkbox
                  :model-value="isTagSelected(tag)"
                  color="primary"
                  hide-details
                />
                <v-list-item-title>{{ getTagName(tag) }}</v-list-item-title>
              </v-list-item-action>
            </v-list-item>

            <div
              ref="sentinel"
              data-test="tags-sentinel"
              style="height: 1px"
            />
          </div>
        </v-list>
        <v-divider v-if="fetchedTags.length > 0" />
        <v-btn
          color="primary"
          text="Manage Tags"
          prepend-icon="mdi-cog"
          class="ma-2"
          :to="{ name: 'Tags' }"
        />
      </div>
    </v-menu>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useIntersectionObserver } from "@vueuse/core";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import useContainersStore from "@/store/modules/containers";
import useDevicesStore from "@/store/modules/devices";
import useTagsStore from "@/store/modules/tags";
import { ITag } from "@/interfaces/ITags";

const props = defineProps<{ variant: "device" | "container" }>();
const containersStore = useContainersStore();
const devicesStore = useDevicesStore();
const tagsStore = useTagsStore();
const snackbar = useSnackbar();
const isMenuOpen = ref(false);
const perPage = ref(10);
const fetchedTags = computed(() => tagsStore.tags);
const selectedTags = computed(() => tagsStore.selectedTags);
const isLoading = ref(false);

const scrollArea = ref<HTMLElement | null>(null);
const sentinel = ref<HTMLElement | null>(null);

const hasMoreTagsToLoad = computed(() => tagsStore.tagCount > fetchedTags.value.length);

const getTagName = (tag: ITag) => typeof tag === "string" ? tag : tag.name;

const getSelectedTagsNames = () => selectedTags.value.map((tag) => getTagName(tag));

const isTagSelected = (selectedTag: ITag) => selectedTags.value.some((tag) => getTagName(tag) === getTagName(selectedTag));

const loadTags = async () => {
  if (isLoading.value) return;
  isLoading.value = true;

  try {
    await tagsStore.fetchTagList({
      filter: "",
      perPage: perPage.value,
    });
  } catch (error) {
    snackbar.showError("Failed to load tags.");
    handleError(error);
  } finally {
    isLoading.value = false;
  }
};

const setFilter = (filter?: string) => {
  if (props.variant === "device") devicesStore.deviceListFilter = filter;
  else containersStore.containerListFilter = filter;
};

const encodeFilter = (tagsNames: string[]) => {
  const filter = [{
    type: "property",
    params: { name: "tags.name", operator: "contains", value: tagsNames },
  }];
  const encodedFilter = Buffer.from(JSON.stringify(filter), "utf-8").toString("base64");
  setFilter(encodedFilter);
};

const selectTag = (tag: ITag) => {
  tagsStore.toggleSelectedTag(tag);

  if (selectedTags.value.length > 0) encodeFilter(getSelectedTagsNames());
  else setFilter();
};

const bumpPerPageAndLoad = async () => {
  if (!hasMoreTagsToLoad.value || isLoading.value) return;
  perPage.value += 10;
  await loadTags();
};

useIntersectionObserver(
  sentinel,
  ([{ isIntersecting }]) => { if (isIntersecting) void bumpPerPageAndLoad(); },
  { root: scrollArea, threshold: 1.0 },
);

onMounted(async () => {
  tagsStore.selectedTags = [];
  await loadTags();
});

defineExpose({ isMenuOpen, loadTags, fetchedTags });
</script>
