<template>
  <div class="mr-4">
    <v-menu location="bottom" scrim eager>
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
      <v-list shaped density="compact">
        <v-list-item-group v-model="selectedTags" multiple>
          <template v-for="(item, i) in getListTags">
            <v-divider v-if="!item" :key="`divider-${i}`"></v-divider>

            <v-list-item
              v-else
              :key="`item-${i}`"
              :value="item"
              active-class="text-deep-purple"
              @click="selectTag(item)"
            >
              <template v-slot:default="{}">
                <div class="d-flex align-center">
                  <v-list-item-action>
                    <v-checkbox
                      :model-value="tagIsSelected(item)"
                      color="deep-purple-accent-3"
                      hide-details
                    />

                    <v-list-item-title v-text="item" />
                  </v-list-item-action>
                </div>
              </template>
            </v-list-item>
          </template>
        </v-list-item-group>
      </v-list>
    </v-menu>
  </div>
</template>

<script lang="ts">
import { computed, defineComponent, onMounted, ref, watch } from "vue";
import { AnyObject } from "yup/lib/object";
import { useStore } from "../../store";

export default defineComponent({
  setup() {
    const store = useStore();

    const prevSelectedLength = ref(0);

    onMounted(() => {
      getTags();
    });

    const getListTags = computed(() => store.getters["tags/list"]);

    const selectedTags = computed(() => store.getters["tags/selected"]);

    const setSelectedTags = (item: any) => {
      store.dispatch("tags/setSelected", item);
    };

    const tagIsSelected = (tag: string) => {
      return selectedTags.value.includes(tag);
    };

    const selectTag = (item: any) => {
      store.dispatch("tags/setSelected", item);
      if (item.length > 0) {
        getDevices(item);
        prevSelectedLength.value = item.length;
      } else if (prevSelectedLength.value === 1 && item.length === 0) {
        fetchDevices();
      }
    };

    const getTags = async () => {
      await store.dispatch("tags/fetch");
    };

    const getDevices = async (item: AnyObject) => {
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
      } catch (error: any) {
        if (error.response.status === 403) {
          store.dispatch("snackbar/showSnackbarErrorAssociation");
        } else {
          store.dispatch("snackbar/showSnackbarErrorDefault");
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
