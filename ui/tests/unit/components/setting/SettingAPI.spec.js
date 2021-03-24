import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import SettingAPI from '@/components/setting/SettingAPI';

describe('SettingAPI', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  const WrapperArray = Array.from({ length: 2 });

  const ownerList = [true, false];

  const namespaces = [
    {
      name: 'namespace1',
      members: [{ id: '4', name: 'user3' }],
      owner: '4',
      tenant_id: 'xxxx',
      setting: {
        webhook: {
          url: '189.129.19.1',
          port: 8909,
          uri: '',
          scheme: 'http',
          active: true,
        },
      },
    },
    {
      name: 'namespace2',
      members: [{ id: '4', name: 'user3' }, { id: '5', name: 'user5' }],
      owner: '5',
      tenant_id: 'xxxx2',
      setting: {
        webhook: {
          url: '189.139.19.1',
          port: 8901,
          uri: '',
          scheme: 'http',
          active: false,
        },
      },
    },
  ];

  const stores = Array.from({ length: 2 }, (_, i) => new Vuex.Store({
    namespaced: true,
    state: {
      namespace: namespaces[i],
      owner: ownerList[i],
    },
    getters: {
      'namespaces/owner': (state) => state.owner,
    },
  }));

  beforeEach(() => {
    WrapperArray.forEach((_, i) => {
      WrapperArray[i] = shallowMount(SettingAPI, {
        store: stores[i],
        localVue,
        stubs: ['fragment'],
      });
    });
  });

  it('Is Vue Instance', () => {
    WrapperArray.forEach((wrapper) => {
      expect(wrapper).toBeTruthy();
    });
  });
  it('Process data in the computed', () => {
    WrapperArray.forEach((wrapper, i) => {
      expect(wrapper.vm.isOwner).toEqual(ownerList[i]);
    });
  });
  it('Render the template for the user owner', () => {
    WrapperArray.forEach((wrapper, i) => {
      expect(wrapper.find('[data-test="webhookOperation"]').exists()).toEqual(ownerList[i]);
    });
  });
});
