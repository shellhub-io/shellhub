import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import Vuetify from 'vuetify';
import NamespaceMemberList from '@/components/app_bar/namespace/NamespaceMemberList';

describe('NamespaceMemberList', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.use(Vuex);

  const namespace = {
    name: 'nsxxx',
    members: [{ username: 'user1', type: 'owner' }, { username: 'user2', type: 'administrator' }, { username: 'user3', type: 'observer' }],
  };

  const headers = [
    {
      text: 'Username',
      value: 'username',
      align: 'start',
      sortable: false,
    },
    {
      text: 'Role',
      value: 'type',
      align: 'center',
      sortable: false,
    },
    {
      text: 'Actions',
      value: 'actions',
      align: 'end',
      sortable: false,
    },
  ];

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      tenant: 'xxxx',
    },
    getters: {
      'auth/tenant': (state) => state.tenant,
    },
    actions: {
      'snackbar/showSnackbarErrorAssociation': () => {},
      'snackbar/showSnackbarErrorAction': () => {},
      'namespaces/get': () => {},
    },
  });

  const wrapper = shallowMount(NamespaceMemberList, {
    store,
    localVue,
    stubs: ['fragment'],
    propsData: { namespace },
    vuetify,
  });

  ///////
  // Component Rendering
  ///////

  it('Is a Vue instance', () => {
    expect(wrapper).toBeTruthy();
  });

  it('Renders the component', () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  ///////
  // Data and Props checking
  ///////

  it('Receive data in props', () => {
    expect(wrapper.vm.namespace).toEqual(namespace);
  });

  it('Compares data with default value', () => {
    expect(wrapper.vm.headers).toEqual(headers);
  });

  it('Proccess data in the computed', () => {
    expect(wrapper.vm.tenant).toEqual('xxxx');
  });

  ///////
  // HTML validation
  ///////

  it('Renders the template with data', () => {
    const dt = wrapper.find('[data-test="dataTable-field"]');
    const dataTableProps = dt.vm.$options.propsData;

    expect(dataTableProps.items).toHaveLength(namespace.members.length);
  });
});
