import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import PublicKeyList from '@/components/public_key/PublicKeyList';

describe('PublicKeyList', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;

  const numberPublicKeys = 2;
  const publicKeys = [
    {
      data: 'BBGVvbmFyZG8=',
      fingerprint: 'b8:26:d5',
      created_at: '2020-11-23T20:59:13.323Z',
      tenant_id: 'xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx',
      name: 'shellhub',
    },
    {
      data: 'AbGVvbmFyZG8=',
      fingerprint: 'b7:25:f8',
      created_at: '2020-11-23T20:59:13.323Z',
      tenant_id: 'xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx',
      name: 'shellhub',
    },
  ];

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      publicKeys,
      numberPublicKeys,
    },
    getters: {
      'publickeys/list': (state) => state.publicKeys,
      'publicKeys/getNumberPublicKeys': (state) => state.numberPublicKeys,
    },
    actions: {
      'publickeys/fetch': () => {
      },
    },
  });

  beforeEach(() => {
    wrapper = shallowMount(PublicKeyList, {
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
  it('Process data in the computed', () => {
    expect(wrapper.vm.getPublicKeys).toEqual(publicKeys);
    expect(wrapper.vm.getNumberPublicKeys).toEqual(numberPublicKeys);
  });
  it('Renders the template with data', () => {
    const dt = wrapper.find('[data-test="dataTable-field"]');
    const dataTableProps = dt.vm.$options.propsData;
    expect(dataTableProps.items).toHaveLength(numberPublicKeys);
  });
});
