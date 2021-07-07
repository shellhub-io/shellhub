import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import DeviceDelete from '@/components/device/DeviceDelete';
import flushPromises from 'flush-promises';
import Vuetify from 'vuetify';

describe('DeviceDelete', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.use(Vuex);

  let wrapper;

  const isOwner = true;
  const uid = 'a582b47a42d';
  const redirect = true;

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      isOwner,
    },
    getters: {
      'namespaces/owner': (state) => state.isOwner,
    },
    actions: {
      'devices/remove': () => {
      },
      'snackbar/showSnackbarSuccessAction': () => {
      },
      'snackbar/showSnackbarErrorAction': () => {
      },
    },
  });

  beforeEach(() => {
    wrapper = mount(DeviceDelete, {
      store,
      localVue,
      stubs: ['fragment'],
      propsData: { uid, redirect },
      vuetify,
    });
  });

  it('Is a Vue instance', () => {
    document.body.setAttribute('data-app', true);
    expect(wrapper).toBeTruthy();
  });
  it('Renders the component', () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
  it('Receive data in props', () => {
    expect(wrapper.vm.uid).toEqual(uid);
    expect(wrapper.vm.redirect).toEqual(redirect);
  });
  it('Compare data with default value', () => {
    expect(wrapper.vm.dialog).toEqual(false);
  });
  it('Show message tooltip to user owner', async (done) => {
    const icons = wrapper.findAll('.v-icon');
    const helpIcon = icons.at(0);
    helpIcon.trigger('mouseenter');
    await wrapper.vm.$nextTick();

    expect(icons.length).toBe(1);
    expect(helpIcon.text()).toEqual('delete');
    requestAnimationFrame(() => {
      expect(wrapper.find('[data-test="tooltipOwner-text"]').text()).toEqual('Remove');
      done();
    });
  });
  it('Renders the template with data', () => {
    expect(wrapper.find('[data-test="deviceDelete-dialog"]').exists()).toEqual(false);
  });
  it('Renders the template with data - dialog is true', async () => {
    wrapper.setData({ dialog: true });
    await flushPromises();

    expect(wrapper.find('[data-test="deviceDelete-dialog"]').exists()).toEqual(true);
  });
});
