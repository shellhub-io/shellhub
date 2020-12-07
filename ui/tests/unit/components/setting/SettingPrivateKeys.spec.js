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
      data: 'BBGVvbmFyZG8=',
    },
    {
      name: 'shellhub',
      data: 'AbGVvbmFyZG8=',
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
  });

  beforeEach(() => {
    wrapper = shallowMount(SettingPrivateKeys, {
      store,
      localVue,
      stubs: ['fragment'],
    });
  });

  it('Is a Vue instance', () => {
    expect(wrapper).toBeTruthy();
  });
  it('Renders the component', () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
  it('Renders the template with data', () => {
    const dt = wrapper.find('[data-test="dataTable-field"]');
    const dataTableProps = dt.vm.$options.propsData;
    expect(dataTableProps.items).toHaveLength(numberPrivateKeys);
  });
  it('Process data in the computed', () => {
    expect(wrapper.vm.getListPrivateKeys).toEqual(privateKeys);
    expect(wrapper.vm.getNumberPrivateKeys).toEqual(numberPrivateKeys);
  });
});
