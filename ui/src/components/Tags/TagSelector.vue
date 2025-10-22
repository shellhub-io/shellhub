<template>
  <div class="mr-4">
    <v-menu v-model="menuOpen" location="bottom" v-bind="$attrs" scrim eager>
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
            :disabled="tags.length === 0"
            @click="loadInitialTags"
          >
            Tags
            <v-icon right>mdi-chevron-down</v-icon>
          </v-btn>
        </v-badge>
      </template>

      <div
        ref="scrollArea"
        class="bg-v-theme-surface"
        style="max-height: 320px; overflow-y: auto;"
      >
        <v-list shaped density="compact">
          <template v-for="(item, i) in tags" :key="`row-${i}`">
            <v-divider v-if="!item" :key="`divider-${i}`" />
            <v-list-item
              v-else
              :key="`item-${i}`"
              :value="item"
              color="primary"
              @click="selectTag(item)"
              data-test="tag-item"
            >
              <template #default>
                <div class="d-flex align-center">
                  <v-list-item-action>
                    <v-checkbox
                      :model-value="tagIsSelected(item)"
                      color="primary"
                      hide-details
                    />
                    <v-list-item-title>{{ getTagName(item) }}</v-list-item-title>
                  </v-list-item-action>
                </div>
              </template>
            </v-list-item>
          </template>

          <div ref="sentinel" data-test="tags-sentinel" style="height: 1px;" />
        </v-list>
      </div>
    </v-menu>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch, nextTick } from "vue";
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

const tenant = computed(() => localStorage.getItem("tenant"));
const menuOpen = ref(false);

const currentPage = ref(1);
const perPage = ref(10);

const fetchedTags = ref<ITag[]>([]);
const tags = computed(() => fetchedTags.value);
const selectedTags = computed<ITag[]>(() => tagsStore.getSelected(props.variant));
const isLoading = ref(false);

const scrollArea = ref<HTMLElement | null>(null);
const sentinel = ref<HTMLElement | null>(null);
let observer: IntersectionObserver | null = null;

const hasMore = computed(() => tagsStore.numberTags > fetchedTags.value.length);

const getTagName = (tag: ITag): string => typeof tag === "string" ? tag : tag.name;

const getSelectedTagNames = (): string[] => selectedTags.value.map((t) => getTagName(t));

const tagIsSelected = (tag: ITag): boolean => selectedTags.value.some(
  (sel) => getTagName(sel) === getTagName(tag),
);

const resetPagination = (): void => {
  currentPage.value = 1;
  perPage.value = 10;
  fetchedTags.value = [];
};

const loadTags = async (): Promise<void> => {
  if (isLoading.value) return;
  isLoading.value = true;

  try {
    await tagsStore.autocomplete({
      tenant: tenant.value || "",
      filter: "",
      page: currentPage.value,
      perPage: perPage.value,
    });

    const newTags = tagsStore.list;
    fetchedTags.value = [...newTags];
  } catch (error) {
    snackbar.showError("Failed to load tags.");
    handleError(error);
  } finally {
    isLoading.value = false;
  }
};

const loadInitialTags = async (): Promise<void> => {
  if (isLoading.value) return;
  resetPagination();
  await loadTags();
};

const setFilter = (filter?: string) => {
  const encoded = filter && filter.length ? filter : undefined;
  if (props.variant === "device") {
    devicesStore.deviceListFilter = encoded;
  } else {
    containersStore.containerListFilter = encoded;
  }
};

const getItems = (tagNames: string[]) => {
  const filter = [
    {
      type: "property",
      params: { name: "tags.name", operator: "contains", value: tagNames },
    },
  ];
  const encodedFilter = Buffer.from(JSON.stringify(filter), "utf-8").toString("base64");
  setFilter(encodedFilter);
};

const selectTag = (item: ITag) => {
  tagsStore.setSelected({ variant: props.variant, tag: item });

  if (selectedTags.value.length > 0) {
    getItems(getSelectedTagNames());
  } else {
    setFilter();
  }
};

const bumpPerPageAndLoad = async () => {
  if (!hasMore.value || isLoading.value) return;
  perPage.value += 10;
  await loadTags();
};

const setupObserver = () => {
  if (observer) {
    observer.disconnect();
    observer = null;
  }
  if (!scrollArea.value || !sentinel.value) return;

  observer = new IntersectionObserver(
    (entries) => {
      const entry = entries[0];
      if (entry?.isIntersecting) {
        bumpPerPageAndLoad();
      }
    },
    { root: scrollArea.value, threshold: 1.0 },
  );

  observer.observe(sentinel.value);
};

const cleanupObserver = () => {
  if (observer) {
    observer.disconnect();
    observer = null;
  }
};

watch(menuOpen, async (open) => {
  if (open) {
    if (fetchedTags.value.length === 0) {
      await loadInitialTags();
    }
    await nextTick();
    setupObserver();
  } else {
    cleanupObserver();
  }
});

onMounted(async () => {
  tagsStore.clearSelected(props.variant);
  await loadTags();
});

onUnmounted(() => {
  cleanupObserver();
});

defineExpose({ menuOpen, loadTags, fetchedTags });
</script>
