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
          @update:search="handleItemsPerPageSearch"
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
import { computed, ref, watch, onMounted } from "vue";
import { useTablePreference, type TableName } from "@/composables/useTablePreference";

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
  tableName?: TableName;
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

const { getItemsPerPage, setItemsPerPage } = useTablePreference();

onMounted(() => {
  if (!props.tableName) return;
  const storedValue = getItemsPerPage(props.tableName);
  if (storedValue !== itemsPerPage.value) itemsPerPage.value = storedValue;
});

watch(itemsPerPage, (newValue, oldValue) => {
  if (props.tableName && newValue !== oldValue && oldValue !== undefined) setItemsPerPage(props.tableName, newValue);
}, { flush: "post" });

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

const clampItemsPerPage = (value: number) => Math.min(100, Math.max(1, value));

const parseItemsPerPageValue = (value: unknown) => {
  if (value === "" || value === null || value === undefined) return null;
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : null;
};

const applyItemsPerPageChange = (value: number) => {
  if (value !== itemsPerPage.value) {
    itemsPerPage.value = value;
    page.value = 1;
  }
  itemsPerPageError.value = null;
};

const updateItemsPerPage = (value: number | string | null) => {
  const parsed = parseItemsPerPageValue(value);
  if (parsed === null) return;

  if (parsed < 1) {
    itemsPerPageError.value = "Minimum is 1";
    return;
  }
  if (parsed > 100) {
    itemsPerPageError.value = "Maximum is 100";
    return;
  }

  applyItemsPerPageChange(parsed);
};

const handleItemsPerPageSearch = (value: string | number | null) => {
  if (value === null || value === undefined) return;
  updateItemsPerPage(value);
};

const readValueFromEvent = (value: unknown) => {
  if (typeof value === "number" || typeof value === "string") return parseItemsPerPageValue(value);
  if (typeof value === "object" && value !== null && "target" in value) {
    const target = value.target as HTMLInputElement | null;
    return target ? parseItemsPerPageValue(target.value) : null;
  }
  return null;
};

const constrainItemsPerPage = (value?: number | string | FocusEvent | KeyboardEvent) => {
  const parsed = readValueFromEvent(value);
  const clamped = clampItemsPerPage(parsed ?? itemsPerPage.value);
  applyItemsPerPageChange(clamped);
};
</script>
