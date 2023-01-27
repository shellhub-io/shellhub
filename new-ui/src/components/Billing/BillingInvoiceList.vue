<template>
  <v-table class="bg-v-theme-surface" v-bind="$attrs">
    <thead>
      <tr>
        <th
          v-for="(head, i) in headers"
          :key="i"
          :class="head.align ? `text-${head.align}` : 'text-center'"
        >
          <span> {{ head.text }}</span>
        </th>
      </tr>
    </thead>
    <tbody v-if="itemsToView.length">
      <tr v-for="(invoice, i) in itemsToView" :key="i">
        <td class="text-center">
          <v-chip v-if="invoice.paid" color="success">
            {{ invoice.status }}
          </v-chip>
          <v-chip
            v-else-if="
              !invoice.paid && invoice.status == 'open' && invoice.attempted
            "
            color="#E53935"
          >
            {{ "payment failed" }}
          </v-chip>

          <v-chip v-else>
            {{ invoice.status }}
          </v-chip>
        </td>

        <td class="text-center">
          {{ formatCurrency(invoice.amountDue, invoice.currency) }}
        </td>

        <td class="text-center">
          {{ unixTimeFormat(invoice.dueDate) }}
        </td>

        <td class="text-center">
          <a v-if="invoice.pdf != '---'" :href="invoice.pdf" target="_blank" rel="noopener noreferrer">
            <v-icon color="#E53935"> mdi-file-pdf-box </v-icon>
          </a>

          <div v-else>
            {{ invoice.pdf }}
          </div>
        </td>

        <td class="text-center">
          <a v-if="invoice.url != '---'" :href="invoice.url" target="_blank" rel="noopener noreferrer">
            <v-icon color="primary"> mdi-credit-card </v-icon>
          </a>

          <div v-else>
            {{ invoice.url }}
          </div>
        </td>
      </tr>
    </tbody>
    <div v-else class="mt-4">
      <p>No data avaliabe</p>
    </div>
  </v-table>
  <v-divider />
  <div class="d-flex w-100 justify-end align-center">
    <span class="text-subtitle-2 mr-4">Items per page:</span>
    <div>
      <v-combobox
        :items="[3, 5, 10]"
        v-model="defaultPerPage"
        outlined
        :update:modelValue="defaultPerPage"
        variant="underlined"
        hide-details
        class="mb-4"
      />
    </div>
    <div class="d-flex align-center">
      <v-btn icon="mdi-chevron-left" variant="plain" @click="previousPage" />
      <span class="text-subtitle-2">{{ page }} of {{ pageQuantity }}</span>
      <v-btn icon="mdi-chevron-right" variant="plain" @click="nextPage" />
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, computed, ref } from "vue";
import formatCurrency from "@/utils/currency";
import unixTimeFormat from "../../utils/timestamp";
import { useStore } from "../../store";

export default defineComponent({
  setup() {
    const store = useStore();
    const defaultPerPage = ref(3);
    const page = ref(1);
    const invoiceList = computed(() => store.getters["billing/getInvoices"]);
    const invoicesLength = computed(
      () => store.getters["billing/getInvoicesLength"],
    );
    const itemsToView = computed(() => {
      const start = (page.value - 1) * defaultPerPage.value;
      const end = start + defaultPerPage.value;
      return invoiceList.value.slice(start, end);
    });
    const pageQuantity = computed(() => Math.ceil(invoicesLength.value / defaultPerPage.value));

    const previousPage = () => {
      if (page.value > 1) {
        page.value--;
      }
    };

    const nextPage = () => {
      if (page.value < pageQuantity.value) {
        page.value++;
      }
    };

    return {
      invoiceList,
      itemsToView,
      defaultPerPage,
      page,
      pageQuantity,
      formatCurrency,
      unixTimeFormat,
      previousPage,
      nextPage,
      headers: [
        {
          text: "Status",
          value: "status",
          align: "center",
          sortable: false,
        },
        {
          text: "Due date",
          value: "dueDate",
          align: "center",
          sortable: false,
        },
        {
          text: "Amount",
          value: "amountDue",
          align: "center",
          sortable: false,
        },
        {
          text: "PDF",
          value: "pdf",
          align: "center",
          sortable: false,
        },
        {
          text: "URL",
          value: "url",
          align: "center",
          sortable: false,
        },
      ],
    };
  },
});
</script>
