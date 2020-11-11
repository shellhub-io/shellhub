import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import NamespaceAdd from '@/components/namespace/NamespaceAdd';

describe('NamespaceAdd', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);
  let wrapper;

  beforeEach(() => {
    wrapper = shallowMount(NamespaceAdd, {
      localVue,
      stubs: ['fragment'],
    });
  });

  it('Is a Vue instance', () => {
    expect(wrapper).toBeTruthy();
  });
  it('Renders the component', () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
  it('Compare data with default value', () => {
    expect(wrapper.vm.dialog).toEqual(false);
    expect(wrapper.vm.namespace).toEqual('');
  });
});
