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
                    <v-list-item-title>{{ item }}</v-list-item-title>
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

const props = defineProps<{ variant: "device" | "container" }>();

const containersStore = useContainersStore();
const devicesStore = useDevicesStore();
const tagsStore = useTagsStore();
const snackbar = useSnackbar();
const tags = computed(() => tagsStore.tags);
const selectedTags = ref<Array<string>>([]);
const tagIsSelected = (tag: string) => selectedTags.value.includes(tag);

const getTags = async () => {
  await tagsStore.fetchTags();
};

const fetchDevices = async (filter?: string) => {
  const fetch = {
    device: () => devicesStore.fetchDeviceList({ filter }),
    container: () => containersStore.fetchContainerList({ filter }),
  }[props.variant];

  await fetch();
};

const getItems = async (item: Array<string>) => {
  const filter = [{
    type: "property",
    params: { name: "tags", operator: "contains", value: item },
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

const selectTag = async (item: string) => {
  if (tagIsSelected(item)) selectedTags.value = selectedTags.value.filter((tag) => tag !== item);
  else selectedTags.value.push(item);

  if (selectedTags.value.length) await getItems(selectedTags.value);
  else await fetchDevices();
};

onMounted(async () => {
  await getTags();
});
</script>
