import Vuex from 'vuex';
import { mount, createLocalVue, config } from '@vue/test-utils';
import NamespaceMenu from '@/components/app_bar/namespace/NamespaceMenu';
import Vuetify from 'vuetify';

config.mocks = {
  $env: {
    isEnterprise: false,
  },
};

describe('NamespaceMenu', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.use(Vuex);

  let wrapper;

  const numberNamespaces = 4;
  const owner = true;
  const inANamespace = true;

  const namespaces = [
    {
      name: 'namespace1',
      owner: 'user1',
      member_names: ['user3', 'user4', 'user5'],
      tenant_id: 'a736a52b-5777-4f92-b0b8-e359bf484713',
    },
    {
      name: 'namespace2',
      owner: 'user1',
      member_names: ['user3', 'user4'],
      tenant_id: 'a736a52b-5777-4f92-b0b8-e359bf484714',
    },
    {
      name: 'namespace3',
      owner: 'user1',
      member_names: ['user6', 'user7', 'user8'],
      tenant_id: 'a736a52b-5777-4f92-b0b8-e359bf484715',
    },
    {
      name: 'namespace4',
      owner: 'user1',
      member_names: ['user6', 'user7'],
      tenant_id: 'a736a52b-5777-4f92-b0b8-e359bf484716',
    },
  ];

  const namespace = {
    name: 'namespace3',
    owner: 'user1',
    member_names: ['user6', 'user7', 'user8'],
    tenant_id: 'a736a52b-5777-4f92-b0b8-e359bf484715',
  };

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      namespace,
      namespaces,
      owner,
    },
    getters: {
      'namespaces/list': (state) => state.namespaces,
      'namespaces/get': (state) => state.namespace,
      'namespaces/owner': (state) => state.owner,
    },
    actions: {
      'namespaces/fetch': () => {
      },
      'namespaces/get': () => {
      },
      'namespaces/switchNamespace': () => {
      },
      'namespaces/setOwnerStatus': () => {
      },
      'snackbar/showSnackbarErrorLoading': () => {
      },
      'snackbar/showSnackbarErrorAssociation': () => {
      },
    },
  });

  beforeEach(() => {
    wrapper = mount(NamespaceMenu, {
      store,
      localVue,
      stubs: ['fragment', 'router-link'],
      propsData: { inANamespace },
      mocks: ['$env'],
      vuetify,
    });
  });

  it('Is a Vue Instance', () => {
    expect(wrapper).toBeTruthy();
  });
  it('Renders the component', () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
  it('Loads menu items with the expected behavior', () => {
    const expectedResuls = [
      {
        height: 150,
        showsList: true,
        available: 3,
      },
      {
        height: 50,
        showsList: true,
        available: 1,
      },
      {
        height: 0,
        showsList: false,
        available: 0,
      },
    ];
    const WrapperArray = [
      wrapper,
      mount(NamespaceMenu, {
        store: new Vuex.Store({
          namespaced: true,
          state: {
            namespace,
            namespaces: [
              {
                name: 'namespace3',
                owner: 'user1',
                member_names: ['user6', 'user7', 'user8'],
                tenant_id: 'a736a52b-5777-4f92-b0b8-e359bf484715',
              },
              {
                name: 'namespace4',
                owner: 'user1',
                member_names: ['user6', 'user7'],
                tenant_id: 'a736a52b-5777-4f92-b0b8-e359bf484716',
              },
            ],
            numberNamespaces,
            owner,
          },
          getters: {
            'namespaces/list': (state) => state.namespaces,
            'namespaces/get': (state) => state.namespace,
            'namespaces/owner': (state) => state.owner,
          },
          actions: {
            'namespaces/fetch': () => {
            },
            'namespaces/get': () => {
            },
            'namespaces/switchNamespace': () => {
            },
            'namespaces/setOwnerStatus': () => {
            },
            'snackbar/showSnackbarErrorLoading': () => {
            },
          },
        }),
        localVue,
        stubs: ['fragment', 'router-link'],
        propsData: { inANamespace },
        mocks: ['$env'],
        vuetify,
      }),
      mount(NamespaceMenu, {
        store: new Vuex.Store({
          namespaced: true,
          state: {
            namespace,
            namespaces: [
              {
                name: 'namespace3',
                owner: 'user1',
                member_names: ['user6', 'user7', 'user8'],
                tenant_id: 'a736a52b-5777-4f92-b0b8-e359bf484715',
              },
            ],
            numberNamespaces,
            owner,
          },
          getters: {
            'namespaces/list': (state) => state.namespaces,
            'namespaces/get': (state) => state.namespace,
            'namespaces/owner': (state) => state.owner,
          },
          actions: {
            'namespaces/fetch': () => {
            },
            'namespaces/get': () => {
            },
            'namespaces/switchNamespace': () => {
            },
            'namespaces/setOwnerStatus': () => {
            },
            'snackbar/showSnackbarErrorLoading': () => {
            },
          },
        }),
        localVue,
        stubs: ['fragment', 'router-link'],
        propsData: { inANamespace },
        mocks: ['$env'],
        vuetify,
      }),
    ];
    expectedResuls.forEach((v, i) => {
      WrapperArray[i].setData({ displayMenu: true });
      expect(WrapperArray[i].vm.namespacesInList).toEqual(v.showsList);
      expect(WrapperArray[i].vm.adaptHeight).toEqual(v.height);
      expect(WrapperArray[i].vm.availableNamespaces.length).toEqual(v.available);
    });
  });
});
