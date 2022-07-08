<template>
  <v-table class="bg-v-theme-surface">
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
    <tbody v-if="invoiceList.length">
      <tr v-for="(invoice, i) in invoiceList" :key="i">
        <td>
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

        <td>
          {{ formatCurrency(invoice.amountDue) }}
        </td>

        <td>
          {{ unixTimeFormat(invoice.dueDate) }}
        </td>

        <td>
          <a v-if="invoice.pdf != '---'" :href="invoice.pdf" target="_blank">
            <v-icon color="#E53935"> mdi-file-pdf-box </v-icon>
          </a>

          <div v-else>
            {{ invoice.pdf }}
          </div>
        </td>

        <td>
          <a v-if="invoice.url != '---'" :href="invoice.url" target="_blank">
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
</template>

<script lang="ts">
import { formatCurrency } from "../../utils/currency";
import unixTimeFormat from "../../utils/timestamp";
import { defineComponent, computed } from "vue";
import { useStore } from "../../store";

export default defineComponent({
  setup() {
    const store = useStore();
    const defaultPerPage = computed(() => store.getters["billing/getPerPage"]);
    const invoiceList = computed(() => store.getters["billing/getInvoices"]);
    const invoicesLength = computed(
      () => store.getters["billing/getInvoicesLength"]
    );

    return {
      invoiceList,
      formatCurrency,
      unixTimeFormat,
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
