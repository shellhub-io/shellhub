import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import PublicKeyFormDialog from '@/components/public_key/PublicKeyFormDialog';

describe('PublicKeyFormDialog', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;

  const publicKey = {
    data: 'AbGVvbmFyZG8=',
    fingerprint: 'b7:25:f8',
    created_at: '2020-11-23T20:59:13.323Z',
    tenant_id: 'xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx',
    name: 'shellhub',
  };
  const createPublicKey = true;

  const store = new Vuex.Store({
    namespaced: true,
    state: {
    },
    getters: {
    },
    actions: {
      'publickeys/post': () => {
      },
      'publickeys/put': () => {
      },
    },
  });

  beforeEach(() => {
    wrapper = shallowMount(PublicKeyFormDialog, {
      store,
      localVue,
      stubs: ['fragment'],
      propsData: { publicKey, createPublicKey },
    });
  });

  it('Is a Vue instance', () => {
    expect(wrapper).toBeTruthy();
  });
  it('Renders the component', () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
  it('Receive data in props', () => {
    expect(wrapper.vm.publicKey).toEqual(publicKey);
    expect(wrapper.vm.createPublicKey).toEqual(createPublicKey);
  });
  it('Compare data with default value', () => {
    expect(wrapper.vm.dialog).toEqual(false);
  });
});
