import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import Vuetify from 'vuetify';
import BillingInvoiceList from '@/components/billing/BillingInvoiceList';

describe('BillingInvoiceList', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.use(Vuex);

  const invoices = [
    {
      paid: false,
      amountDue: 3040,
      dueDate: 1614983421,
      status: 'open',
      pdf: 'test.pdf',
      url: 'inv_url',
    },
    {
      paid: true,
      amountDue: 2030,
      dueDate: 1614983421,
      status: 'paid',
      pdf: '---',
      url: '---',
    },
  ];

  const headers = [
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
  ];

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      invoices,
      invoicesLength: invoices.length,
      defaultPerPage: 3,
    },
    getters: {
      'billing/getInvoices': (state) => state.invoices,
      'billing/defaultPerPage': (state) => state.defaultPerPage,
      'billing/getInvoicesLength': (state) => state.invoiceinvoicesLength,
    },
    actions: {
      'billing/getPagination': () => {},
    },
  });

  const wrapper = shallowMount(BillingInvoiceList, {
    localVue,
    store,
    vuetify,
    stubs: ['fragment'],
    propsData: { invoices },
  });

  ///////
  // Component Rendering
  //////

  it('Is a Vue instance', () => {
    expect(wrapper).toBeTruthy();
  });
  it('Renders the component', () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  ///////
  // Data and props checking
  //////

  it('Compares data with default value', () => {
    expect(wrapper.vm.headers).toStrictEqual(headers);
  });
  it('Process data in the computed', () => {
    expect(wrapper.vm.invoiceList).toBe(invoices);
  });

  ///////
  // HTML validation
  //////

  it('Renders the template with data', () => {
    const dt = wrapper.find('[data-test="dataTable-field"]');
    const dataTableProps = dt.vm.$options.propsData;
    dataTableProps.items.forEach((item, i) => {
      expect(item).toStrictEqual(invoices[i]);
    });

    expect(dataTableProps.items).toHaveLength(invoices.length);
  });
});
