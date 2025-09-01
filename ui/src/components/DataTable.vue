<template>
  <div>
    <v-table :class="isAdmin ? 'bg-surface' : 'bg-background border rounded'" class="text-center">
      <thead class="bg-v-theme-background">
        <tr>
          <th v-for="(header, i) in headers" :key="i" class="text-center">
            <span
              v-if="header.sortable"
              @click="$emit('update:sort', header.value)"
              @keypress.enter="$emit('update:sort', header.value)"
              tabindex="0"
              class="cursor-pointer text-decoration-underline"
            >
              {{ header.text }}
              <v-tooltip activator="parent" anchor="top">Sort by {{ header.text }}</v-tooltip>
            </span>
            <span v-else> {{ header.text }}</span>
          </th>
        </tr>
      </thead>
      <tbody v-if="items.length">
        <slot name="rows" />
      </tbody>
      <tbody v-else class="pa-4 text-subtitle-2">
        <tr>
          <td :colspan="headers.length" class="pa-4 text-subtitle-2 text-center">
            No data available
          </td>
        </tr>
      </tbody>
    </v-table>
    <v-divider v-if="isAdmin" />
    <v-progress-linear v-if="loading" indeterminate alt="Data table loading" />
    <div class="d-flex w-100 justify-end align-center" v-if="itemsPerPageOptions?.length">
      <span class="text-subtitle-2 mr-4">Items per page:</span>
      <div>
        <v-combobox
          :items="itemsPerPageOptions"
          v-model="itemsPerPage"
          @update:model-value="goToFirstPage"
          outlined
          variant="underlined"
          hide-details
          class="mb-4"
        />
      </div>
      <div class="d-flex align-center">
        <v-btn icon="mdi-chevron-left" variant="plain" @click="page--" :disabled="page <= 1" />
        <span class="text-subtitle-2">{{ page }} of {{ pageQuantity }}</span>
        <v-btn
          icon="mdi-chevron-right"
          variant="plain"
          @click="page++"
          :disabled="pageQuantity <= 1 || page === pageQuantity" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, inject } from "vue";

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
const isAdmin: boolean = inject("isAdmin", false);

const goToFirstPage = () => { page.value = 1; };
</script>
