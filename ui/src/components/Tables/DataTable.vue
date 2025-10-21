<template>
  <div data-test="datatable-root">
    <v-table class="bg-background border rounded text-center" data-test="datatable">
      <thead class="bg-v-theme-background">
        <tr>
          <th v-for="(header, i) in headers" :key="i" class="text-center" :data-test="`th-${header.value}`">
            <span
              v-if="header.sortable"
              @click="$emit('update:sort', header.value)"
              @keypress.enter="$emit('update:sort', header.value)"
              tabindex="0"
              class="cursor-pointer text-decoration-underline"
              :data-test="`sort-${header.value}`"
            >
              {{ header.text }}
              <v-tooltip activator="parent" anchor="top">Sort by {{ header.text }}</v-tooltip>
            </span>
            <span v-else data-test="th-label">{{ header.text }}</span>
          </th>
        </tr>
      </thead>

      <tbody v-if="items.length" data-test="tbody-has-items">
        <slot name="rows" />
      </tbody>

      <tbody v-else class="pa-4 text-subtitle-2" data-test="tbody-empty">
        <tr>
          <td
            :colspan="headers.length"
            class="pa-4 text-subtitle-2 text-center"
            data-test="empty-state"
          >
            No data available
          </td>
        </tr>
      </tbody>
    </v-table>

    <v-progress-linear
      v-if="loading"
      indeterminate
      alt="Data table loading"
      data-test="loading"
    />

    <div
      class="d-flex w-100 justify-end align-center"
      v-if="itemsPerPageOptions?.length"
      data-test="pager"
    >
      <span class="text-subtitle-2 mr-4" data-test="ipp-label">Items per page:</span>

      <div>
        <v-combobox
          :items="itemsPerPageOptions"
          v-model="itemsPerPage"
          @update:model-value="goToFirstPage"
          outlined
          variant="underlined"
          hide-details
          class="mb-4"
          data-test="ipp-combo"
        />
      </div>

      <div class="d-flex align-center" data-test="pager-controls">
        <v-btn
          icon="mdi-chevron-left"
          variant="plain"
          @click="page--"
          :disabled="page <= 1"
          data-test="pager-prev"
        />
        <span class="text-subtitle-2" data-test="pager-text">{{ page }} of {{ pageQuantity }}</span>
        <v-btn
          icon="mdi-chevron-right"
          variant="plain"
          @click="page++"
          :disabled="pageQuantity <= 1 || page === pageQuantity"
          data-test="pager-next"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";

type Header = {
  text: string;
  value: string;
  sortable?: boolean;
};

const props = defineProps<{
  headers: Header[];
  items: object[];
  totalCount: number;
  loading: boolean;
  itemsPerPageOptions?: number[];
}>();

defineEmits(["update:sort"]);

const page = defineModel<number>("page", { required: true, type: Number });
const itemsPerPage = defineModel("itemsPerPage", { required: true, type: Number });
const pageQuantity = computed(() => Math.ceil(props.totalCount / itemsPerPage.value) || 1);
const goToFirstPage = () => { page.value = 1; };
</script>
