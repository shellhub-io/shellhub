import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome';
import BillingIcon from '@/components/billing/BillingIcon';

describe('BillingIcon', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  localVue.component('FontAwesomeIcon', FontAwesomeIcon);

  let wrapper;

  const iconName = 'cc-amex';
  const defaultIcon = 'credit-card';

  const cardIcon = {
    amex: 'cc-amex',
    dinersClub: 'cc-diners-club',
    discover: 'cc-discover',
    jcb: 'cc-jcb',
    mastercard: 'cc-mastercard',
    visa: 'cc-visa',
  };

  beforeEach(() => {
    wrapper = shallowMount(BillingIcon, {
      localVue,
      stubs: ['fragment'],
      propsData: { iconName },
    });
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
  // Data and Props checking
  //////

  it('Receive data in props', () => {
    expect(wrapper.vm.iconName).toBe(iconName);
  });
  it('Compare data with default value', () => {
    expect(wrapper.vm.cardIcon).toEqual(cardIcon);
  });

  //////
  // HTML validation
  //////

  it('Renders the template with data', () => {
    //////
    // In this case, the default icon is tested.
    //////

    wrapper = shallowMount(BillingIcon, {
      localVue,
      stubs: ['fragment'],
      propsData: { iconName: defaultIcon },
    });

    expect(wrapper.find('[data-test="default-icon"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="type-icon"]').exists()).toBe(false);
  });

  //////
  // In this case, the other icons are tested.
  //////

  Object.keys(cardIcon).forEach((iconKey) => {
    wrapper = shallowMount(BillingIcon, {
      localVue,
      stubs: ['fragment'],
      propsData: { iconName: iconKey },
    });

    expect(wrapper.find('[data-test="default-icon"]').exists()).toBe(false);
    expect(wrapper.find('[data-test="type-icon"]').exists()).toBe(true);
  });
});
