<template>
  <fragment>
    <div class="5">
      <v-data-table
        class="elevation-0"
        :headers="headers"
        :items="invoiceList"
        :items-per-page="defaultPerPage"
        :footer-props="{'items-per-page-options': [3, 5, 10]}"
        :server-items-length="invoicesLength"
        data-test="dataTable-field"
        :options.sync="pagination"
      >
        <template #[`item.status`]="{ item }">
          <v-chip
            v-if="item.paid"
            color="success"
          >
            {{ item.status }}
          </v-chip>

          <v-chip
            v-else-if="!item.paid && item.status=='open' && item.attempted"
            color="#E53935"
          >
            {{ 'payment failed' }}
          </v-chip>

          <v-chip v-else>
            {{ item.status }}
          </v-chip>
        </template>

        <template #[`item.amountDue`]="{ item }">
          {{ item.amountDue | formatCurrency }}
        </template>

        <template #[`item.dueDate`]="{ item }">
          {{ item.dueDate | unixTimeFormat }}
        </template>

        <template #[`item.pdf`]=" {item} ">
          <a
            v-if="item.pdf!='---'"
            :href="item.pdf"
            target="_blank"
          >
            <v-icon color="#E53935">
              mdi-file-pdf-box
            </v-icon>
          </a>

          <div v-else>
            {{ item.pdf }}
          </div>
        </template>

        <template #[`item.url`]=" {item} ">
          <a
            v-if="item.url!='---'"
            :href="item.url"
            target="_blank"
          >
            <v-icon color="primary">
              mdi-credit-card
            </v-icon>
          </a>

          <div v-else>
            {{ item.url }}
          </div>
        </template>
      </v-data-table>
    </div>
  </fragment>
</template>

<script>

import formatCurrency from '@/components/filter/currency';
import simpleFormat from '@/components/filter/date';
import unixTimeFormat from '@/helpers/timestamp';

export default {
  name: 'BillingInvoiceList',

  filters: {
    formatCurrency,
    simpleFormat,
    unixTimeFormat,
  },

  data() {
    return {
      pagination: {},
      headers: [
        {
          text: 'Status',
          value: 'status',
          align: 'center',
          sortable: false,
        },
        {
          text: 'Due date',
          value: 'dueDate',
          align: 'center',
          sortable: false,
        },
        {
          text: 'Amount',
          value: 'amountDue',
          align: 'center',
          sortable: false,
        },
        {
          text: 'PDF',
          value: 'pdf',
          align: 'center',
          sortable: false,
        },
        {
          text: 'URL',
          value: 'url',
          align: 'center',
          sortable: false,
        },
      ],
    };
  },

  computed: {
    defaultPerPage() {
      return this.$store.getters['billing/getPerPage'];
    },

    invoiceList() {
      return this.$store.getters['billing/getInvoices'];
    },

    invoicesLength() {
      return this.$store.getters['billing/getInvoicesLength'];
    },
  },

  watch: {
    pagination: {
      handler() {
        this.$store.dispatch('billing/getPagination', { perPage: this.pagination.itemsPerPage, page: this.pagination.page });
      },
      deep: true,
    },
  },
};

</script>
