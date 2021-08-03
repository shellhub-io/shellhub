import Vuex from 'vuex';
import { mount, createLocalVue, config } from '@vue/test-utils';
import NamespaceMenu from '@/components/app_bar/namespace/NamespaceMenu';
import Vuetify from 'vuetify';

config.mocks = {
  $env: {
    isEnterprise: true,
  },
};

describe('NamespaceMenu', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.use(Vuex);

  let wrapper;
  let wrapperArray;

  const numberNamespaces = 4;
  const owner = true;
  const inANamespace = true;
  const isMobile = false;

  const namespace = {
    name: 'namespace3',
    owner: 'user1',
    member_names: ['user6', 'user7', 'user8'],
    tenant_id: 'e359bf484715',
  };

  const expectedResuls = [
    {
      height: 100,
      showsList: true,
      available: 2,
    },
    {
      height: 50,
      showsList: true,
      available: 1,
    },
  ];

  const namespaces = [
    {
      name: 'namespace1',
      owner: 'user1',
      member_names: ['user3', 'user4', 'user5'],
      tenant_id: 'xxxxxxxx',
    },
    {
      name: 'namespace2',
      owner: 'user1',
      member_names: ['user3', 'user4'],
      tenant_id: 'xxxxxxxx',
    },
    {
      name: 'namespace3',
      owner: 'user1',
      member_names: ['user6', 'user7', 'user8'],
      tenant_id: 'xxxxxxxx',
    },
    {
      name: 'namespace4',
      owner: 'user1',
      member_names: ['user6', 'user7'],
      tenant_id: 'xxxxxxxx',
    },
  ];

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      namespace,
      namespaces,
      owner,
      isMobile,
    },
    getters: {
      'namespaces/list': (state) => state.namespaces,
      'namespaces/get': (state) => state.namespace,
      'namespaces/owner': (state) => state.owner,
      'mobile/isMobile': (state) => state.isMobile,
    },
    actions: {
      'namespaces/fetch': () => {},
      'namespaces/get': () => {},
      'namespaces/switchNamespace': () => {},
      'namespaces/setOwnerStatus': () => {},
      'snackbar/showSnackbarErrorLoading': () => {},
      'snackbar/showSnackbarErrorAssociation': () => {},
    },
  });

  const storeWithTwoNamespaces = new Vuex.Store({
    namespaced: true,
    state: {
      namespace,
      namespaces: [
        {
          name: 'namespace3',
          owner: 'user1',
          member_names: ['user6', 'user7', 'user8'],
          tenant_id: 'xxxxxxxx',
        },
        {
          name: 'namespace4',
          owner: 'user1',
          member_names: ['user6', 'user7'],
          tenant_id: 'xxxxxxxx',
        },
      ],
      numberNamespaces,
      owner,
      isMobile,
    },
    getters: {
      'namespaces/list': (state) => state.namespaces,
      'namespaces/get': (state) => state.namespace,
      'namespaces/owner': (state) => state.owner,
      'mobile/isMobile': (state) => state.isMobile,
    },
    actions: {
      'namespaces/fetch': () => {},
      'namespaces/get': () => {},
      'namespaces/switchNamespace': () => {},
      'namespaces/setOwnerStatus': () => {},
      'snackbar/showSnackbarErrorLoading': () => {},
    },
  });

  const storeWithOneNamespace = new Vuex.Store({
    namespaced: true,
    state: {
      namespace,
      namespaces: [
        {
          name: 'namespace3',
          owner: 'user1',
          member_names: ['user6', 'user7', 'user8'],
          tenant_id: 'xxxxxxxx',
        },
      ],
      numberNamespaces,
      owner,
      isMobile,
    },
    getters: {
      'namespaces/list': (state) => state.namespaces,
      'namespaces/get': (state) => state.namespace,
      'namespaces/owner': (state) => state.owner,
      'mobile/isMobile': (state) => state.isMobile,
    },
    actions: {
      'namespaces/fetch': () => {},
      'namespaces/get': () => {},
      'namespaces/switchNamespace': () => {},
      'namespaces/setOwnerStatus': () => {},
      'snackbar/showSnackbarErrorLoading': () => {},
    },
  });

  ///////
  // In this case, check owner fields rendering in enterprise version
  // and mobile of the template.
  ///////

  describe('', () => {
    beforeEach(() => {
      wrapper = mount(NamespaceMenu, {
        store,
        localVue,
        stubs: ['fragment', 'router-link'],
        propsData: { inANamespace: !inANamespace },
        mocks: ['$env'],
        vuetify,
      });

      localStorage.setItem('tenant', namespace.tenant_id);
    });

    ///////
    // Component Rendering
    //////

    it('Is a Vue Instance', () => {
      expect(wrapper).toBeTruthy();
    });
    it('Renders the component', () => {
      expect(wrapper.html()).toMatchSnapshot();
    });

    ///////
    // Data and Props checking
    //////

    it('Receives data in props', () => {
      expect(wrapper.vm.inANamespace).toEqual(!inANamespace);
    });
    it('Compare data with default value', () => {
      expect(wrapper.vm.model).toEqual(true);
      expect(wrapper.vm.dialog).toEqual(false);
      expect(wrapper.vm.displayMenu).toEqual(false);
      expect(wrapper.vm.defaultSize).toEqual(24);
      expect(wrapper.vm.first).toEqual(false);
    });
    it('Process data in the computed', () => {
      expect(wrapper.vm.adaptHeight).toEqual(150);
      expect(wrapper.vm.isOwner).toEqual(owner);
      expect(wrapper.vm.namespace).toEqual(namespace);
      expect(wrapper.vm.namespaces).toEqual(namespaces);
      expect(wrapper.vm.namespacesInList).toEqual(true);
      expect(wrapper.vm.availableNamespaces).toEqual(namespaces);
      expect(wrapper.vm.loggedInNamespace).toEqual(!inANamespace);
      expect(wrapper.vm.tenant).toEqual(namespace.tenant_id);
      expect(wrapper.vm.isEnterprise).toEqual(true);
      expect(wrapper.vm.isMobile).toEqual(false);
    });

    //////
    // HTML validation
    //////

    it('Renders the template with components', () => {
      expect(wrapper.find('[data-test="namespaceAdd-component"]').exists()).toEqual(true);
    });

    it('Renders the template with data', () => {
      expect(wrapper.find('[data-test="add-btn"]').exists()).toEqual(true);
      expect(wrapper.find('[data-test="namespaceMenu-menu"]').exists()).toEqual(false);
    });
  });

  ///////
  // In this case, check owner fields rendering in enterprise version
  // of the template.
  ///////

  describe('Button', () => {
    beforeEach(() => {
      wrapper = mount(NamespaceMenu, {
        store,
        localVue,
        stubs: ['fragment', 'router-link'],
        propsData: { inANamespace },
        mocks: ['$env'],
        vuetify,
      });

      localStorage.setItem('tenant', namespace.tenant_id);
    });

    ///////
    // Component Rendering
    //////

    it('Is a Vue Instance', () => {
      expect(wrapper).toBeTruthy();
    });
    it('Renders the component', () => {
      expect(wrapper.html()).toMatchSnapshot();
    });

    ///////
    // Data and Props checking
    //////

    it('Receives data in props', () => {
      expect(wrapper.vm.inANamespace).toEqual(inANamespace);
    });
    it('Compare data with default value', () => {
      expect(wrapper.vm.model).toEqual(true);
      expect(wrapper.vm.dialog).toEqual(false);
      expect(wrapper.vm.displayMenu).toEqual(false);
      expect(wrapper.vm.defaultSize).toEqual(24);
      expect(wrapper.vm.first).toEqual(false);
    });
    it('Process data in the computed', () => {
      expect(wrapper.vm.adaptHeight).toEqual(150);
      expect(wrapper.vm.isOwner).toEqual(owner);
      expect(wrapper.vm.namespace).toEqual(namespace);
      expect(wrapper.vm.namespaces).toEqual(namespaces);
      expect(wrapper.vm.namespacesInList).toEqual(true);
      expect(wrapper.vm.availableNamespaces).toEqual(namespaces);
      expect(wrapper.vm.loggedInNamespace).toEqual(inANamespace);
      expect(wrapper.vm.tenant).toEqual(namespace.tenant_id);
      expect(wrapper.vm.isEnterprise).toEqual(true);
      expect(wrapper.vm.isMobile).toEqual(false);
    });

    //////
    // HTML validation
    //////

    it('Renders the template with components', () => {
      expect(wrapper.find('[data-test="namespaceAdd-component"]').exists()).toEqual(true);
    });

    it('Renders the template with data', () => {
      expect(wrapper.find('[data-test="add-btn"]').exists()).toEqual(false);
      expect(wrapper.find('[data-test="namespaceMenu-menu"]').exists()).toEqual(true);
    });
  });

  ///////
  // In this case, check owner fields rendering in open version
  // of the template.
  ///////

  describe('Menu', () => {
    beforeEach(() => {
      wrapper = mount(NamespaceMenu, {
        store,
        localVue,
        stubs: ['fragment', 'router-link'],
        propsData: { inANamespace },
        mocks: {
          $env: {
            isEnterprise: false,
          },
        },
        vuetify,
      });

      localStorage.setItem('tenant', namespace.tenant_id);
    });

    ///////
    // Component Rendering
    //////

    it('Is a Vue Instance', () => {
      expect(wrapper).toBeTruthy();
    });
    it('Renders the component', () => {
      expect(wrapper.html()).toMatchSnapshot();
    });

    ///////
    // Data and Props checking
    //////

    it('Receives data in props', () => {
      expect(wrapper.vm.inANamespace).toEqual(inANamespace);
    });
    it('Compare data with default value', () => {
      expect(wrapper.vm.model).toEqual(true);
      expect(wrapper.vm.dialog).toEqual(false);
      expect(wrapper.vm.displayMenu).toEqual(false);
      expect(wrapper.vm.defaultSize).toEqual(24);
      expect(wrapper.vm.first).toEqual(false);
    });
    it('Process data in the computed', () => {
      expect(wrapper.vm.adaptHeight).toEqual(150);
      expect(wrapper.vm.isOwner).toEqual(owner);
      expect(wrapper.vm.namespace).toEqual(namespace);
      expect(wrapper.vm.namespaces).toEqual(namespaces);
      expect(wrapper.vm.namespacesInList).toEqual(true);
      expect(wrapper.vm.availableNamespaces).toEqual(namespaces);
      expect(wrapper.vm.loggedInNamespace).toEqual(inANamespace);
      expect(wrapper.vm.tenant).toEqual(namespace.tenant_id);
      expect(wrapper.vm.isEnterprise).toEqual(false);
      expect(wrapper.vm.isMobile).toEqual(false);
    });

    //////
    // HTML validation
    //////

    it('Renders the template with components', () => {
      expect(wrapper.find('[data-test="namespaceAdd-component"]').exists()).toEqual(true);
    });

    it('Renders the template with data', () => {
      expect(wrapper.find('[data-test="add-btn"]').exists()).toEqual(false);
      expect(wrapper.find('[data-test="namespaceMenu-menu"]').exists()).toEqual(true);
    });
  });

  ///////
  // In this case, check owner fields rendering in enterprise version
  // and different mobile sizes of the template.
  ///////

  describe('Different mobile sizes', () => {
    beforeEach(() => {
      wrapperArray = [
        mount(NamespaceMenu, {
          store: storeWithTwoNamespaces,
          localVue,
          stubs: ['fragment', 'router-link'],
          propsData: { inANamespace },
          mocks: ['$env'],
          vuetify,
        }),

        mount(NamespaceMenu, {
          store: storeWithOneNamespace,
          localVue,
          stubs: ['fragment', 'router-link'],
          propsData: { inANamespace },
          mocks: ['$env'],
          vuetify,
        }),
      ];
    });

    ///////
    // Component Rendering
    //////

    it('Is a Vue Instance', () => {
      expectedResuls.forEach((v, i) => {
        expect(wrapperArray[i]).toBeTruthy();
      });
    });
    it('Renders the component', () => {
      expectedResuls.forEach((v, i) => {
        expect(wrapperArray[i].html()).toMatchSnapshot();
      });
    });

    //////
    // HTML validation
    //////

    it('Renders the template with components', () => {
      expectedResuls.forEach((v, i) => {
        expect(wrapperArray[i].find('[data-test="namespaceAdd-component"]').exists()).toEqual(true);
      });
    });
    it('Renders the template with data', () => {
      expectedResuls.forEach((v, i) => {
        wrapperArray[i].setData({ displayMenu: true });
        expect(wrapperArray[i].vm.namespacesInList).toEqual(v.showsList);
        expect(wrapperArray[i].vm.adaptHeight).toEqual(v.height);
        expect(wrapperArray[i].vm.availableNamespaces.length).toEqual(v.available);

        expect(wrapperArray[i].find('[data-test="add-btn"]').exists()).toEqual(false);
        expect(wrapperArray[i].find('[data-test="namespaceMenu-menu"]').exists()).toEqual(true);
      });
    });
  });
});
