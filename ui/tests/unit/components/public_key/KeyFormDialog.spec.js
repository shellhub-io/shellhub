import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import KeyFormDialog from '@/components/public_key/KeyFormDialog';

describe('KeyFormDialog', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;

  const keyObject = {
    data: 'AbGVvbmFyZG8=',
    fingerprint: 'b7:25:f8',
    created_at: '2020-11-23T20:59:13.323Z',
    tenant_id: 'xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx',
    name: 'shellhub',
  };
  const createKey = true;
  const action = 'Public';

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
    wrapper = shallowMount(KeyFormDialog, {
      store,
      localVue,
      stubs: ['fragment'],
      propsData: { keyObject, createKey, action },
    });
  });

  it('Is a Vue instance', () => {
    expect(wrapper).toBeTruthy();
  });
  it('Renders the component', () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
  it('Receive data in props', () => {
    expect(wrapper.vm.keyObject).toEqual(keyObject);
    expect(wrapper.vm.createKey).toEqual(createKey);
    expect(wrapper.vm.action).toEqual(action);
  });
  it('Compare data with default value', () => {
    expect(wrapper.vm.dialog).toEqual(false);
  });
});
