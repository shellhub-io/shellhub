import { shallowMount, createLocalVue } from '@vue/test-utils';
import Vuetify from 'vuetify';
import BillingPaymentList from '@/components/billing/BillingPaymentList';

describe('BillingPaymentList', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();

  const pms = [
    {
      brand: 'visa',
      expYear: 2024,
      default: true,
      expMonth: 4,
      last4: '4042',
      id: 'pm_1JzQ80KJsksFHO6pREJA5TrK',
    },
    {
      brand: 'visa',
      expYear: 2028,
      default: false,
      expMonth: 4,
      last4: '4042',
      id: 'pm_1JzQ80KJsksFHO6pREJA5TrG',
    },
    {
      brand: 'visa',
      expYear: 2029,
      default: false,
      expMonth: 4,
      last4: '4042',
      id: 'pm_1JzQ80KJsksFHO6pREJA5TrF',
    },
  ];

  const headers = [
    {
      text: 'Brand',
      value: 'brand',
      align: 'center',
      sortable: false,
    },
    {
      text: 'Exp. Date',
      value: 'expdate',
      align: 'center',
      sortable: false,
    },
    {
      text: 'Ends with',
      value: 'last4',
      align: 'center',
      sortable: false,
    },
    {
      text: 'Actions',
      value: 'actions',
      align: 'center',
      sortable: false,
    },
  ];

  const wrapper = shallowMount(BillingPaymentList, {
    localVue,
    vuetify,
    stubs: ['fragment'],
    propsData: { cards: pms },
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
    expect(wrapper.vm.paymentList).toBe(pms);
  });
  it('Process data in props', () => {
    expect(wrapper.vm.cards).toBe(pms);
  });

  ///////
  // HTML validation
  //////

  it('Renders the template with data', () => {
    const dt = wrapper.find('[data-test="dataTable-field"]');
    const dataTableProps = dt.vm.$options.propsData;
    dataTableProps.items.forEach((item, i) => {
      expect(item).toStrictEqual(pms[i]);
    });

    expect(dataTableProps.items).toHaveLength(pms.length);
  });
});
