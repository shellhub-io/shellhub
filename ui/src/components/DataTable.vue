<template>
  <div class="bg-v-theme-surface">
    <v-table class="bg-v-theme-surface">
      <thead>
        <tr>
          <th v-for="(head, i) in headers" :key="i" :class="head.align ? `text-${head.align}` : 'text-center'">
            <span
              v-if="head.sortable"
              @click="$emit('clickSortableIcon', head.value)"
              @keypress.enter="$emit('clickSortableIcon', head.value)"
              tabindex="0"
              class="hover"
            >
              {{ head.text }}
              <v-tooltip activator="parent" anchor="top">Sort by {{ head.text }}</v-tooltip>
            </span>
            <span v-else> {{ head.text }}</span>
          </th>
        </tr>
      </thead>
      <tbody v-if="items.length">
        <slot name="rows" />
      </tbody>
      <div v-else class="pa-4 text-subtitle-2">
        <p>No data available</p>
      </div>
    </v-table>
    <v-divider />
    <v-progress-linear v-if="loading" indeterminate alt="Data table loading" />
    <div class="d-flex w-100 justify-end align-center" v-if="!itemSelectorDisable == true">
      <span class="text-subtitle-2 mr-4">Items per page:</span>
      <div>
        <v-combobox
          :items="comboboxOptions"
          v-model="itemsPerPageRef"
          outlined
          :update:modelValue="$emit('changeItemsPerPage', itemsPerPageRef)"
          variant="underlined"
          hide-details
          class="mb-4"
        />
      </div>
      <div class="d-flex align-center">
        <v-btn icon="mdi-chevron-left" variant="plain" @click="$emit('clickPreviousPage')" :disabled="pageQuantity <= 1" />
        <span class="text-subtitle-2">{{ actualPage }} of {{ pageQuantity }}</span>
        <v-btn
          icon="mdi-chevron-right"
          variant="plain"
          @click="$emit('clickNextPage')"
          :disabled="pageQuantity <= 1 || actualPage == pageQuantity" />
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { computed, defineComponent, PropType, toRefs, ref } from "vue";

type HeaderItem = {
  text: string;
  value: string;
  sortable?: boolean;
  align?: "center" | "left" | "right";
};

type UserTable = {
  name: string;
  email: string;
  username: string;
  namespaces: string;
};

export default defineComponent({
  props: {
    headers: {
      type: Array as PropType<HeaderItem[]>,
      default: () => [],
      required: true,
    },
    items: {
      type: Array,
      default: () => [] as PropType<UserTable[]>,
      required: false,
    },
    itemsPerPage: {
      type: Number,
      required: true,
    },
    itemSelectorDisable: {
      type: Boolean,
      required: false,
    },
    comboboxOptions: {
      type: Array as PropType<number[]>,
      required: false,
      default: () => [10, 20, 50, 100],
    },
    loading: {
      type: Boolean,
      default: false,
      required: false,
    },
    actualPage: {
      type: Number,
      default: 1,
    },
    totalCount: {
      type: Number,
      required: true,
    },
    nextPage: {
      type: Function as PropType<() => void>,
      default: Function as PropType<() => {}>,
    },
    previousPage: {
      type: Function as PropType<() => void>,
      default: Function as PropType<() => {}>,
    },
  },
  emits: ["changeItemsPerPage", "clickNextPage", "clickPreviousPage", "clickSortableIcon"],
  setup(props) {
    const { itemsPerPage, totalCount } = toRefs(props);

    const itemsPerPageRef = ref(itemsPerPage.value);
    const pageQuantity = computed(() => Math.ceil(totalCount.value / itemsPerPageRef.value));

    return {
      itemsPerPageRef,
      pageQuantity,
    };
  },
});
</script>

<style scoped>
.hover:hover {
  cursor: pointer;
  text-decoration: underline;
}
</style>
