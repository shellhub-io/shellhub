import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import NamespaceDelete from '@/components/app_bar/namespace/NamespaceDelete';
import flushPromises from 'flush-promises';
import Vuetify from 'vuetify';

describe('NamespaceDelete', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  document.body.setAttribute('data-app', true);
  localVue.use(Vuex);

  let wrapper;

  const tenant = 'xxxxxx';

  const namespace = {
    name: 'namespace3',
    owner: 'user1',
    member_names: ['user6', 'user7', 'user8'],
    tenant_id: 'a736a52b-5777-4f92-b0b8-e359bf484715',
  };

  const text = `This action cannot be undone. This will permanently delete the
         ${namespace.name} and its related data.`;

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      namespace,
    },
    getters: {
      'namespaces/get': (state) => state.namespace,
    },
    actions: {
      'namespaces/remove': () => {
      },
      'auth/logout': () => {
      },
      'snackbar/showSnackbarErrorLoading': () => {
      },
      'snackbar/showSnackbarSuccessAction': () => {
      },
    },
  });

  beforeEach(() => {
    wrapper = mount(NamespaceDelete, {
      store,
      localVue,
      stubs: ['fragment'],
      propsData: { nsTenant: tenant },
      vuetify,
    });
  });

  it('Is a Vue instance', () => {
    expect(wrapper).toBeTruthy();
  });
  it('Renders the component', () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
  it('Receives data in props', () => {
    expect(wrapper.vm.nsTenant).toEqual(tenant);
  });
  it('Compare data with the default', () => {
    expect(wrapper.vm.name).toEqual(namespace.name);
    expect(wrapper.vm.dialog).toEqual(false);
  });
  it('Renders the template with data', async () => {
    expect(wrapper.find('[data-test="namespaceDelete-dialog"]').exists()).toEqual(false);
  });
  it('Renders the template with data - dialog is true', async () => {
    wrapper.setData({ dialog: true });
    await flushPromises();

    expect(wrapper.find('[data-test="content-text"]').text()).toEqual(text);
  });
});
