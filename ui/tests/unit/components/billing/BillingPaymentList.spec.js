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

  const heading = [
    {
      id: 'brand',
      title: 'Brand',
    },
    {
      id: 'expdate',
      title: 'Exp. Date',
    },
    {
      id: 'last4',
      title: 'Ends with',
    },
    {
      id: 'actions',
      title: 'Actions',
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
  // Data checking
  //////

  it('Compares data with default value', () => {
    expect(wrapper.vm.heading).toStrictEqual(heading);
  });
  it('Process data in the computed', () => {
    expect(wrapper.vm.paymentList).toBe(pms);
  });

  //////
  // HTML validation
  //////

  heading.forEach((item) => {
    it(`Renders the template heading ${item.id} with data`, () => {
      expect(wrapper.find(`[data-test="${item.id}-div"]`).text()).toBe(item.title);
    });
  });

  pms.forEach((pm, i) => {
    it(`Renders the template for payment row ${i} with data`, () => {
      expect(wrapper.find(`[data-test="icon-${pm.id}-component"]`).exists()).toBe(true);
      expect(wrapper.find(`[data-test="exp-date-${pm.id}-col"]`).exists()).toBe(true);
      expect(wrapper.find(`[data-test="exp-date-${pm.id}-col"]`).text()).toBe(`${pm.expMonth}/${pm.expYear}`);
      expect(wrapper.find(`[data-test="last4-${pm.id}-col"]`).text()).toBe(`${pm.last4}`);
      if (pm.default) {
        expect(wrapper.find(`[data-test="default-${pm.id}-div"]`).exists()).toBe(true);
        expect(wrapper.find(`[data-test="default-${pm.id}-div"]`).text()).toBe('Default');
        expect(wrapper.find(`[data-test="actions-${pm.id}-div"]`).exists()).toBe(false);
      } else {
        expect(wrapper.find(`[data-test="actions-${pm.id}-div"]`).exists()).toBe(true);
        expect(wrapper.find(`[data-test="actions-${pm.id}-div"]`).text()).toContain('Make default');
        expect(wrapper.find(`[data-test="actions-${pm.id}-div"]`).text()).toContain('Remove');
        expect(wrapper.find(`[data-test="default-${pm.id}-div"]`).exists()).toBe(false);
      }
    });
  });
});
