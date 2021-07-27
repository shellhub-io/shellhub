import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import PublicKeyList from '@/components/public_key/PublicKeyList';

describe('PublicKeyList', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;

  const status = true;
  const numberPublicKeys = 2;

  const publicKeys = [
    {
      data: 'BBGVvbmF',
      fingerprint: '00:00:00',
      created_at: '2020-11-23T20:59:13.323Z',
      tenant_id: 'xxxxxxxx',
      name: 'shellhub',
    },
    {
      data: 'AbGVvbmF',
      fingerprint: '00:00:00',
      created_at: '2020-11-23T20:59:13.323Z',
      tenant_id: 'xxxxxxxx',
      name: 'shellhub',
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
      value: 'fingerprint',
      align: 'center',
    },
    {
      text: 'Hostname',
      value: 'hostname',
      align: 'center',
    },
    {
      text: 'Created At',
      value: 'created_at',
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
      publicKeys,
      numberPublicKeys,
      status,
    },
    getters: {
      'publickeys/list': (state) => state.publicKeys,
      'publickeys/getNumberPublicKeys': (state) => state.numberPublicKeys,
      'boxs/getStatus': (state) => state.status,
    },
    actions: {
      'publickeys/fetch': () => {},
      'snackbar/showSnackbarErrorLoading': () => {},
      'boxs/setStatus': () => {},
    },
  });

  beforeEach(() => {
    wrapper = shallowMount(PublicKeyList, {
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
    expect(wrapper.vm.headers).toEqual(headers);
  });
  it('Process data in the computed', () => {
    expect(wrapper.vm.getPublicKeys).toEqual(publicKeys);
    expect(wrapper.vm.getNumberPublicKeys).toEqual(numberPublicKeys);
  });

  //////
  // HTML validation
  //////

  it('Renders the template with data', () => {
    const dt = wrapper.find('[data-test="publickeyList-dataTable"]');
    const dataTableProps = dt.vm.$options.propsData;

    expect(dataTableProps.items).toHaveLength(numberPublicKeys);
  });
});
