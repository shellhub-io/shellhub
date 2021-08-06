import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import Vuetify from 'vuetify';
import SettingOwnerInfo from '@/components/setting/SettingOwnerInfo';

describe('SettingOwnerInfo', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);
  const vuetify = new Vuetify();

  let wrapper;

  const isOwner = true;

  const namespace = {
    name: 'namespace1',
    owner: '124',
    members: [{ id: '124', name: 'user4' }, { id: '123', name: 'user1' }, { id: '125', name: 'user5' }],
    tenant_id: 'a736a52b-5777-4f92-b0b8-e359bf484713',
  };

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      namespace,
    },
    getters: {
      'namespaces/get': (state) => state.namespace,
    },
    actions: {
    },
  });

  ///////
  // In this case, when the user owns the namespace and the focus of
  // the test is to check if the message no exists.
  ///////

  describe('Owner is true', () => {
    beforeEach(() => {
      wrapper = shallowMount(SettingOwnerInfo, {
        store,
        localVue,
        propsData: { isOwner },
        vuetify,
        stubs: ['fragment'],
      });
    });

    ///////
    // Component Rendering
    //////

    it('Is a Vue intance', () => {
      expect(wrapper).toBeTruthy();
    });
    it('Renders the component', () => {
      expect(wrapper.html()).toMatchSnapshot();
    });

    //////
    // HTML validation
    //////

    it('Renders the template with data', () => {
      expect(wrapper.find('[data-test="message-div"]').exists()).toBe(false);
    });
  });

  ///////
  // In this case, when the user owns the namespace and the focus of
  // the test is to check if the message exists.
  ///////

  describe('Owner is false', () => {
    beforeEach(() => {
      wrapper = shallowMount(SettingOwnerInfo, {
        store,
        localVue,
        propsData: { isOwner: !isOwner },
        vuetify,
        stubs: ['fragment'],
      });
    });

    ///////
    // Component Rendering
    //////

    it('Is a Vue intance', () => {
      expect(wrapper).toBeTruthy();
    });
    it('Renders the component', () => {
      expect(wrapper.html()).toMatchSnapshot();
    });

    //////
    // HTML validation
    //////

    it('Renders the template with data', () => {
      expect(wrapper.find('[data-test="message-div"]').exists()).toBe(true);

      const namespaceOwnerMessage = `Contact ${namespace.members[0].name} user for more information.`;
      expect(wrapper.find('[data-test=contactUser-p]').text()).toEqual(namespaceOwnerMessage);
    });
  });
});
