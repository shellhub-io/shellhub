import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import SettingPrivateKeys from '@/components/setting/SettingPrivateKeys';

describe('SettingPrivateKeys', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;

  const numberPrivateKeys = 2;

  const privateKeys = [
    {
      name: 'shellhub',
      data: 'BBGVvbmF',
    },
    {
      name: 'shellhub',
      data: 'AbGVvbmF',
    },
  ];

  const headers = [
    {
      text: 'Name',
      value: 'name',
      align: 'center',
    },
    {
      text: 'Fingerprint',
      value: 'data',
      align: 'center',
    },
    {
      text: 'Actions',
      value: 'actions',
      align: 'center',
    },
  ];

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      privateKeys,
      numberPrivateKeys,
    },
    getters: {
      'privatekeys/list': (state) => state.privateKeys,
      'privatekeys/getNumberPrivateKeys': (state) => state.numberPrivateKeys,
    },
    actions: {
    },
  });

  ///////
  // In this case, the first time the user enters the private keys
  // tab, a dialog appears with a certain message.
  ///////

  describe('Dialog is true', () => {
    beforeEach(() => {
      wrapper = shallowMount(SettingPrivateKeys, {
        store,
        localVue,
        stubs: ['fragment'],
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

    it('Compare data with default value', () => {
      expect(wrapper.vm.dialog).toEqual(true);
      expect(wrapper.vm.privatekeyPrivacyPolicy).toEqual(false);
      expect(wrapper.vm.headers).toEqual(headers);
    });
    it('Process data in the computed', () => {
      expect(wrapper.vm.getListPrivateKeys).toEqual(privateKeys);
      expect(wrapper.vm.getNumberPrivateKeys).toEqual(numberPrivateKeys);
    });

    //////
    // HTML validation
    //////

    it('Renders the template with data', () => {
      expect(wrapper.find('[data-test="gotIt-btn"]').exists()).toBe(true);
    });
  });

  ///////
  // In this case, verify data in data table.
  ///////

  describe('Dialog is false', () => {
    beforeEach(() => {
      wrapper = shallowMount(SettingPrivateKeys, {
        store,
        localVue,
        stubs: ['fragment'],
      });

      wrapper.setData({ dialog: false });
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

    it('Compare data with default value', () => {
      expect(wrapper.vm.dialog).toEqual(false);
      expect(wrapper.vm.privatekeyPrivacyPolicy).toEqual(false);
      expect(wrapper.vm.headers).toEqual(headers);
    });
    it('Process data in the computed', () => {
      expect(wrapper.vm.getListPrivateKeys).toEqual(privateKeys);
      expect(wrapper.vm.getNumberPrivateKeys).toEqual(numberPrivateKeys);
    });

    //////
    // HTML validation
    //////

    it('Renders the template with components', () => {
      expect(wrapper.find('[data-test="privateKeyFormDialogFirst-component"]').exists()).toBe(true);
    });
    it('Renders the template with data', () => {
      const dt = wrapper.find('[data-test="dataTable-field"]');
      const dataTableProps = dt.vm.$options.propsData;

      expect(dataTableProps.items).toHaveLength(numberPrivateKeys);
    });
  });
});
