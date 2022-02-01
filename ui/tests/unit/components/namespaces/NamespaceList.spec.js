import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import Vuetify from 'vuetify';
import NamespaceList from '@/components/namespace/NamespaceList';

describe('Namespace', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.use(Vuex);

  let wrapper;

  const namespace = {
    name: 'namespace3',
    owner: 'user1',
    member_names: ['user6', 'user7', 'user8'],
    tenant_id: 'e359bf484715',
  };

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
      tenant_id: 'xxxxxxxy',
    },
  ];

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      namespace,
      namespaces,
    },
    getters: {
      'namespaces/list': (state) => state.namespaces,
      'namespaces/get': (state) => state.namespace,
    },
    actions: {
      'namespaces/switchNamespace': () => {},
      'snackbar/showSnackbarErrorLoading': () => {},
    },
  });

  ///////
  // In this case, check owner fields rendering in enterprise version.
  ///////

  describe('Enterprise version', () => {
    beforeEach(() => {
      wrapper = shallowMount(NamespaceList, {
        store,
        localVue,
        stubs: ['fragment'],
        vuetify,
      });
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

    it('Process data in the computed', () => {
      expect(wrapper.vm.namespace).toEqual(namespace);
      expect(wrapper.vm.namespaces).toEqual(namespaces.filter((el) => el.name !== namespace.name));
    });

    //////
    // HTML validation
    //////

    it('Renders the template with data', async () => {
      const namespacesLocal = namespaces.filter((el) => el.name !== namespace.name);

      Object.keys(namespacesLocal).forEach((item) => {
        expect(wrapper.find(`[data-test="${namespacesLocal[item].name}-namespace"]`).text()).toEqual(namespacesLocal[item].name);
      });
    });
  });

  ///////
  // In this case, check owner fields rendering in open version
  // of the template.
  ///////

  describe('Open version', () => {
    beforeEach(() => {
      wrapper = shallowMount(NamespaceList, {
        store,
        localVue,
        stubs: ['fragment', 'router-link'],
        vuetify,
      });

      jest.spyOn(Storage.prototype, 'getItem').mockReturnValue(namespace.tenant_id);
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

    it('Process data in the computed', () => {
      expect(wrapper.vm.namespace).toEqual(namespace);
      expect(wrapper.vm.namespaces).toEqual(namespaces.filter((el) => el.name !== namespace.name));
    });

    //////
    // HTML validation
    //////

    it('Renders the template with data', async () => {
      const namespacesLocal = namespaces.filter((el) => el.name !== namespace.name);

      Object.keys(namespacesLocal).forEach((item) => {
        expect(wrapper.find(`[data-test="${namespacesLocal[item].name}-namespace"]`).text()).toEqual(namespacesLocal[item].name);
      });
    });
  });
});
