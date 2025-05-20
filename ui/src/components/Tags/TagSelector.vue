<template>
  <div class="mr-4">
    <v-menu
      v-model="menuOpen"
      location="bottom"
      v-bind="$attrs"
      scrim
      eager
      open-on-click
      :close-on-content-click="false"
    >
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
            :disabled="tags.length == 0"
          >
            Tags
            <v-icon right> mdi-chevron-down </v-icon>
          </v-btn>
        </v-badge>
      </template>
      <v-list shaped density="compact" class="bg-v-theme-surface" max-height="600">
        <template v-for="(item, i) in tags">
          <v-divider v-if="!item" :key="`divider-${i}`" />
          <v-list-item
            v-else
            :key="`item-${i}`"
            :value="item"
            active-color="primary"
            @click="selectTag(item)"
          >
            <template v-slot:default="{}">
              <div class="d-flex justify-start" action>
                <v-list-item>
                  <v-checkbox
                    color="primary"
                    hide-details
                  />
                  <v-list-item-title><v-chip>{{ item.name }} </v-chip></v-list-item-title>
                </v-list-item>
              </div>
            </template>
          </v-list-item>
        </template>
        <div class="d-flex justify-center" v-if="hasMore">
          <v-btn
            @click="loadMoreTags"
            variant="text"
            color="primary"
            data-test="load-more-tags-btn"
            :disabled="!hasMore"
          >
            Load More
          </v-btn>
        </div>
      </v-list>
    </v-menu>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, PropType } from "vue";
import axios, { AxiosError } from "axios";
import { useStore } from "@/store";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import { Tags } from "@/interfaces/ITags";

const props = defineProps({
  variant: {
    type: String as PropType<"device" | "container">,
    required: true,
  },
});

const tenant = computed(() => localStorage.getItem("tenant"));
const page = ref(1);
const perPage = ref(10);
const fetchedTags = ref<Array<Tags>>([]);
const tags = computed(() => fetchedTags.value);
const store = useStore();
const snackbar = useSnackbar();
const prevSelectedLength = ref(0);
const selectedTags = computed<Array<Tags>>(() => store.getters["tags/selected"](props.variant));
const hasMore = ref(true);
const menuOpen = ref(false);

const getTags = async () => {
  try {
    await store.dispatch("tags/autocomplete", {
      tenant: tenant.value,
      page: page.value,
      perPage: perPage.value,
    });

    const newTags = store.getters["tags/list"];

    if (newTags.length < perPage.value) hasMore.value = false;
    else page.value += 1;

    fetchedTags.value = [...fetchedTags.value, ...newTags];
  } catch (error) {
    console.error("Failed to load tags", error);
  }
};

const loadMoreTags = async () => {
  await getTags();
  menuOpen.value = true;
};

const getItems = async (item: Array<Tags>) => {
  let encodedFilter: string | null = null;

  const filter = [
    {
      type: "property",
      params: { name: "tags", operator: "contains", value: item },
    },
  ];
  encodedFilter = btoa(JSON.stringify(filter));

  await store.dispatch("tag/setFilter", encodedFilter);
  switch (props.variant) {
    case "device":
      await store.dispatch("devices/setFilter", encodedFilter);
      break;
    case "container":
      await store.dispatch("container/setFilter", encodedFilter);
      break;
    default:
      break;
  }

  try {
    switch (props.variant) {
      case "device":
        await store.dispatch("devices/refresh", encodedFilter);
        break;
      case "container":
        await store.dispatch("container/refresh", encodedFilter);
        break;
      default:
        break;
    }
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

const fetchDevices = async () => {
  const data = {
    perPage:
      props.variant === "device"
        ? store.getters["devices/getPerPage"]
        : store.getters["container/getPerPage"],
    page:
      props.variant === "device"
        ? store.getters["devices/getPage"]
        : store.getters["container/getPage"],
    status: "accepted",
    search: null,
    filter: "",
    sortStatusField: null,
  };

  switch (props.variant) {
    case "device":
      await store.dispatch("devices/fetch", data);
      break;
    case "container":
      await store.dispatch("container/fetch", data);
      break;
    default:
      break;
  }
};

const selectTag = async (item: Tags) => {
  store.commit("tags/setSelected", { variant: props.variant, tag: item });

  if (selectedTags.value.length > 0) {
    await getItems(selectedTags.value);
    prevSelectedLength.value = selectedTags.value.length;
  } else {
    await fetchDevices();
  }
};

onMounted(() => {
  store.commit("tags/clearSelected", props.variant);
  getTags();
});

defineExpose({ loadMoreTags, fetchedTags, menuOpen });
</script>
