import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import SettingWebhook from '@/components/setting/SettingWebhook';
import Vuetify from 'vuetify';

describe('SettingWebhook', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);
  const vuetify = new Vuetify();

  const WrapperArray = Array.from({ length: 2 });

  const namespaces = [{
    name: 'namespace1',
    owner: 'user1',
    members: [{ name: 'user1' }, { name: 'user2' }],
    tenant_id: 'a736a52b-5777-4f92-b0b8-e359bf484731',
    settings: {
      webhook: {
        url: 'http://189.192.19.3:8080',
        active: true,
      },
    },
  },
  {
    name: 'namespace2',
    owner: 'user3',
    members: [{ name: 'user3' }, { name: 'user4' }],
    tenant_id: 'a736a52b-5777-4f92-b0b8-e359bf484799',
    settings: {
      session_record: false,
      webhook: {
        url: 'http://domain.test:7777',
        active: true,
      },
    },
  },
  ];

  const stores = Array.from({ length: 2 }, (_, i) => new Vuex.Store({
    namespaced: true,
    state: {
      webhook: namespaces[i].settings.webhook,
    },
    getters: {
      'namespaces/webhook': (state) => state.webhook,
      'namespaces/webhookActive': (state) => state.active,
    },
    action: {
      'namespace/updateWebhook': () => {
      },
    },
  }));

  beforeEach(() => {
    localStorage.setItem('tenant', namespaces[0].tenant_id);
    WrapperArray.forEach((_, i) => {
      WrapperArray[i] = mount(SettingWebhook, {
        store: stores[i],
        localVue,
        stubs: ['fragment'],
        vuetify,
      });
    });
  });

  it('Is a Vue instance', () => {
    WrapperArray.forEach((wrapper) => {
      expect(wrapper).toBeTruthy();
    });
  });
  it('Renders the component', () => {
    WrapperArray.forEach((wrapper) => {
      expect(wrapper.html()).toMatchSnapshot();
    });
  });
  it('Renders text-fields elements', () => {
    WrapperArray.forEach((wrapper) => {
      expect(wrapper.find('[data-test="field-url"]').exists()).toBe(true);
    });
  });
  it('Returns data back to the state value', () => {
    WrapperArray.forEach((wrapper, i) => {
      wrapper.setData({ webhookUrlField: '' });
      wrapper.vm.setWebhookData();
      expect(wrapper.vm.webhookUrlField).toBe(stores[i].getters['namespaces/webhookUrl']);
    });
  });
});
