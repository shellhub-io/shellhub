import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import SettingOwnerInfo from '@/components/setting/SettingOwnerInfo';
import Vuetify from 'vuetify';

describe('SettingOwnerInfo', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);
  const vuetify = new Vuetify();

  const WrapperArray = Array.from({ length: 2 });

  const isOwnerArray = [
    true,
    false,
  ];

  const namespaces = [
    {
      name: 'namespace1',
      owner: '124',
      members: [{ id: '123', name: 'user1' }, { id: '124', name: 'user4' }, { id: '125', name: 'user5' }],
      tenant_id: 'a736a52b-5777-4f92-b0b8-e359bf484713',
    },
    {
      name: 'namespace2',
      owner: '24',
      members: [{ id: '23', name: 'user1' }, { id: '24', name: 'user4' }, { id: '25', name: 'user5' }],
      tenant_id: 'a736a52b-5777-4f92-b0b8-e359bf484715',
    },
  ];

  const stores = Array.from({ length: 2 }, (_, i) => new Vuex.Store({
    namespaced: true,
    state: {
      namespace: namespaces[i],
    },
    getters: {
      'namespaces/get': (state) => state.namespace,
    },
  }));

  beforeEach(() => {
    WrapperArray.forEach(async (_, i) => {
      WrapperArray[i] = shallowMount(SettingOwnerInfo, {
        store: stores[i],
        localVue,
        propsData: { isOwner: isOwnerArray[i] },
        vuetify,
        stubs: ['fragment'],
      });
    });
  });

  it('Is a Vue intance', () => {
    WrapperArray.forEach((wrapper) => {
      expect(wrapper).toBeTruthy();
    });
  });
  it('Renders the component', () => {
    WrapperArray.forEach((wrapper) => {
      expect(wrapper.html()).toMatchSnapshot();
    });
  });
  it('Show and hides component according to owner', () => {
    WrapperArray.forEach((wrapper, i) => {
      expect(wrapper.find('[data-test="notTheOwner"]').exists()).toBe(!isOwnerArray[i]);
      expect(wrapper.find('[data-test="namespaceOwnerMessage"]').exists()).toBe(!isOwnerArray[i]);
    });
  });
});
