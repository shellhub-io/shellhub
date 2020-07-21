import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import DetailsDevice from '@/views/DetailsDevice';

describe('Terminal', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;

  beforeEach(() => {
    wrapper = shallowMount(DetailsDevice, {
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
});
