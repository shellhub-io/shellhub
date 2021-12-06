import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import Vuetify from 'vuetify';
import NamespaceMemberList from '@/components/app_bar/namespace/NamespaceMemberList';

describe('NamespaceMemberList', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.use(Vuex);

  const heading = [
    {
      id: 'name',
      title: 'Username',
    },
    {
      id: 'role',
      title: 'Role',
    },
    {
      id: 'actions',
      title: 'Actions',
    },
  ];

  const namespace = {
    name: 'nsxxx',
    members: [{ username: 'user1', type: 'owner' }, { username: 'user2', type: 'administrator' }, { username: 'user3', type: 'observer' }],
  };

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
    expect(wrapper.vm.heading).toEqual(heading);
  });

  it('Proccess data in the computed', () => {
    expect(wrapper.vm.tenant).toEqual('xxxx');
  });

  ///////
  // HTML validation
  ///////

  it('Renders the template with data', () => {
    const { members } = namespace;
    members.forEach((m, i) => {
      expect(wrapper.find(`[data-test="${m.username}-list"]`).text()).toBe(members[i].username);
      expect(wrapper.find(`[data-test="${m.type}-list"]`).text()).toBe(members[i].type);
      expect(wrapper.find(`[data-test="${m.username}-actions-list"]`).exists()).toBe(true);
    });
    heading.forEach((col) => {
      expect(wrapper.find(`[data-test="${col.title}-title"]`).text()).toBe(col.title);
    });
  });
});
