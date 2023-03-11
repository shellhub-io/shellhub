<!-- eslint-disable vue/no-v-text-v-html-on-component -->
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
            :disabled="getListTags.length == 0"
            @click="getTags"
          >
            Tags
            <v-icon right> mdi-chevron-down </v-icon>
          </v-btn>
        </v-badge>
      </template>
      <v-list shaped density="compact" class="bg-v-theme-surface">
        <div>
          <template v-for="(item, i) in getListTags">
            <v-divider v-if="!item" :key="`divider-${i}`" />

            <v-list-item
              v-else
              :key="`item-${i}`"
              :value="item"
              active-color="primary"
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

                    <v-list-item-title v-text="item" />
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

<script lang="ts">
import { computed, defineComponent, onMounted, ref } from "vue";
import axios, { AxiosError } from "axios";
import { useStore } from "../../store";
import handleError from "@/utils/handleError";

export default defineComponent({
  inheritAttrs: true,
  setup() {
    const store = useStore();

    const prevSelectedLength = ref(0);

    const getListTags = computed(() => store.getters["tags/list"]);

    const selectedTags = computed<Array<string>>(() => store.getters["tags/selected"]);

    const setSelectedTags = (item: Array<string>) => {
      store.dispatch("tags/setSelected", item);
    };

    const tagIsSelected = (tag: string) => selectedTags.value.includes(tag);

    const getTags = async () => {
      await store.dispatch("tags/fetch");
    };

    const getDevices = async (item: Array<string>) => {
      let encodedFilter : string | null = null;

      const filter = [
        {
          type: "property",
          params: { name: "tags", operator: "contains", value: item },
        },
      ];
      encodedFilter = btoa(JSON.stringify(filter));

      await store.dispatch("devices/setFilter", encodedFilter);

      try {
        store.dispatch("devices/refresh");
      } catch (error: unknown) {
        if (axios.isAxiosError(error)) {
          const axiosError = error as AxiosError;
          if (axiosError.response?.status === 403) {
            store.dispatch("snackbar/showSnackbarErrorAssociation");
            handleError(error);
          }
        } else {
          store.dispatch("snackbar/showSnackbarErrorDefault");
          handleError(error);
        }
      }
    };

    const fetchDevices = async () => {
      const data = {
        perPage: store.getters["devices/getPerPage"],
        page: store.getters["devices/getPage"],
        status: "accepted",
        search: null,
        filter: "",
        sortStatusField: null,
      };

      await store.dispatch("devices/fetch", data);
    };

    const selectTag = async (item: string) => {
      store.dispatch("tags/setSelected", item);
      if (selectedTags.value.length > 0) {
        await getDevices(selectedTags.value);
        prevSelectedLength.value = selectedTags.value.length;
      } else if (prevSelectedLength.value === 1 && selectedTags.value.length === 0) {
        await fetchDevices();
      }

      if (selectedTags.value.length === 0) {
        await store.dispatch("tags/clearSelectedTags");
        await fetchDevices();
      }
    };

    onMounted(() => {
      getTags();
    });

    return {
      prevSelectedLength,
      selectedTags,
      setSelectedTags,
      getListTags,
      getTags,
      tagIsSelected,
      selectTag,
    };
  },
});
</script>
