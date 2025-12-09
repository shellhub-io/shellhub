<template>
  <div data-test="datatable-root">
    <v-table
      class="bg-background border rounded text-center"
      data-test="datatable"
    >
      <thead class="bg-v-theme-background">
        <tr>
          <th
            v-for="(header, i) in headers"
            :key="i"
            class="text-center"
            :data-test="`th-${header.value}`"
          >
            <span
              v-if="header.sortable"
              tabindex="0"
              class="cursor-pointer text-decoration-underline"
              :data-test="`sort-${header.value}`"
              @click="$emit('update:sort', header.value)"
              @keypress.enter="$emit('update:sort', header.value)"
            >
              {{ header.text }}
              <v-tooltip
                activator="parent"
                anchor="top"
              >Sort by {{ header.text }}</v-tooltip>
            </span>
            <span
              v-else
              data-test="th-label"
            >{{ header.text }}</span>
          </th>
        </tr>
      </thead>

      <tbody
        v-if="items.length"
        data-test="tbody-has-items"
      >
        <slot name="rows" />
      </tbody>

      <tbody
        v-else
        class="pa-4 text-subtitle-2"
        data-test="tbody-empty"
      >
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

    <v-row
      v-if="itemsPerPageOptions?.length"
      class="w-100 pt-3"
      align="center"
      justify="end"
      data-test="pager"
    >
      <v-col
        cols="auto"
        class="pa-0"
      >
        <span
          class="text-subtitle-2 mr-4"
          data-test="ipp-label"
        >
          Items per page:
        </span>
      </v-col>

      <v-col
        cols="auto"
        class="pa-0"
      >
        <v-combobox
          :model-value="itemsPerPage"
          :items="itemsPerPageOptions"
          :error-messages="itemsPerPageError"
          variant="underlined"
          type="number"
          hide-spin-buttons
          hide-details="auto"
          hide-no-data
          class="mb-4 mr-1 w-100"
          data-test="ipp-combo"
          @update:model-value="updateItemsPerPage"
          @blur="constrainItemsPerPage"
          @keydown.enter="constrainItemsPerPage"
          @keydown="blockNonNumeric"
          @paste.prevent
        />
      </v-col>

      <v-col
        cols="auto"
        class="pa-0"
      >
        <div
          class="d-flex align-center"
          data-test="pager-controls"
        >
          <v-btn
            icon="mdi-chevron-left"
            variant="plain"
            :disabled="page <= 1"
            data-test="pager-prev"
            @click="page--"
          />

          <span
            class="text-subtitle-2"
            data-test="pager-text"
          >
            {{ page }} of {{ pageQuantity }}
          </span>

          <v-btn
            icon="mdi-chevron-right"
            variant="plain"
            :disabled="pageQuantity <= 1 || page === pageQuantity"
            data-test="pager-next"
            @click="page++"
          />
        </div>
      </v-col>
    </v-row>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";

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

const page = defineModel<number>("page", {
  required: true,
  type: Number,
});

const itemsPerPage = defineModel<number>("itemsPerPage", {
  required: true,
  type: Number,
});

const itemsPerPageError = ref<string | null>(null);
const pageQuantity = computed(() => Math.ceil(props.totalCount / itemsPerPage.value) || 1);

const blockNonNumeric = (e: KeyboardEvent) => {
  const allowedKeys = [
    "Backspace",
    "Delete",
    "ArrowLeft",
    "ArrowRight",
    "Tab",
    "Home",
    "End",
    "Enter",
  ];

  if (allowedKeys.includes(e.key)) return;
  if (/^[0-9]$/.test(e.key)) return;

  e.preventDefault();
};

const updateItemsPerPage = (value: number) => {
  if (value < 1) {
    itemsPerPageError.value = "Minimum is 1";
    return;
  }
  if (value > 100) {
    itemsPerPageError.value = "Maximum is 100";
    return;
  }

  itemsPerPageError.value = null;
  if (value !== itemsPerPage.value) itemsPerPage.value = value;
};

const constrainItemsPerPage = (value: number) => {
  const clamped = Math.max(1, Math.min(100, value));
  itemsPerPage.value = clamped;
  page.value = 1;
  itemsPerPageError.value = null;
};
</script>
