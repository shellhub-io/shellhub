<template>
  <div class="mr-4">
    <v-menu location="bottom" v-bind="$attrs" scrim eager>
      <template v-slot:activator="{ props }">
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
            @click="getTags"
          >
            Tags
            <v-icon right> mdi-chevron-down </v-icon>
          </v-btn>
        </v-badge>
      </template>
      <v-list shaped density="compact" class="bg-v-theme-surface">
        <div>
          <template v-for="(item, i) in tags">
            <v-divider v-if="!item" :key="`divider-${i}`" />

            <v-list-item
              v-else
              :key="`item-${i}`"
              :value="item"
              color="primary"
              @click="selectTag(item)"
            >
              <template v-slot:default="{}">
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
        </div>
      </v-list>
    </v-menu>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import axios, { AxiosError } from "axios";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import useContainersStore from "@/store/modules/containers";
import useDevicesStore from "@/store/modules/devices";
import useTagsStore from "@/store/modules/tags";
import { Tags } from "@/interfaces/ITags";

const props = defineProps<{ variant: "device" | "container" }>();

const containersStore = useContainersStore();
const devicesStore = useDevicesStore();
const tagsStore = useTagsStore();
const snackbar = useSnackbar();
const tenant = computed(() => localStorage.getItem("tenant"));
const page = ref(1);
const perPage = ref(10);
const fetchedTags = ref<Tags[]>([]);
const tags = computed(() => fetchedTags.value);
const prevSelectedLength = ref(0);
const selectedTags = computed<Tags[]>(() => tagsStore.getSelected(props.variant));
const hasMore = ref(true);

const getTagName = (tag: Tags): string => typeof tag === "string" ? tag : tag.name;

const getSelectedTagNames = (): string[] => selectedTags.value.map((tag) => getTagName(tag));

const tagIsSelected = (tag: Tags): boolean => {
  const tagName = getTagName(tag);
  return selectedTags.value.some((selectedTag) => getTagName(selectedTag) === tagName);
};

const getTags = async (): Promise<void> => {
  try {
    await tagsStore.autocomplete({
      tenant: tenant.value || "",
      filter: "",
      page: page.value,
      perPage: perPage.value,
    });

    const newTags = tagsStore.list;

    if (newTags.length < perPage.value) hasMore.value = false;
    else page.value += 1;

    fetchedTags.value = [...fetchedTags.value, ...newTags];
  } catch (error) {
    snackbar.showError("Failed to load tags.");
    console.error("Failed to load tags", error);
  }
};

const fetchDevices = async (filter?: string): Promise<void> => {
  const fetch = {
    device: () => devicesStore.fetchDeviceList({ filter }),
    container: () => containersStore.fetchContainerList({ filter }),
  }[props.variant];

  await fetch();
};

const getItems = async (tagNames: string[]): Promise<void> => {
  const filter = [{
    type: "property",
    params: { name: "tags", operator: "contains", value: tagNames },
  }];

  const encodedFilter = btoa(JSON.stringify(filter));

  try {
    await fetchDevices(encodedFilter);
  } catch (error: unknown) {
    if (axios.isAxiosError(error)) {
      const axiosError = error as AxiosError;
      if (axiosError.response?.status === 403) {
        snackbar.showError("You do not have permission to perform this action.");
        handleError(error);
      }
    } else {
      snackbar.showError("Failed to load items.");
      handleError(error);
    }
  }
};

const selectTag = async (item: Tags): Promise<void> => {
  tagsStore.setSelected({ variant: props.variant, tag: item });

  if (selectedTags.value.length > 0) {
    const selectedTagNames = getSelectedTagNames();
    await getItems(selectedTagNames);
    prevSelectedLength.value = selectedTags.value.length;
  } else {
    await fetchDevices();
  }
};

onMounted(async () => {
  tagsStore.clearSelected(props.variant);
  await getTags();
});
</script>
