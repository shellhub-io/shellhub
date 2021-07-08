import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import DeviceRename from '@/components/device/DeviceRename';
import Vuetify from 'vuetify';

describe('DeviceRename', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.use(Vuex);

  let wrapper;

  const isOwner = true;
  const uid = 'a582b47a42d';
  const name = '39-5e-2a';

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      isOwner,
    },
    getters: {
      'namespaces/owner': (state) => state.isOwner,
      'devices/get': (state) => state.device,
    },
    actions: {
      'devices/rename': () => {
      },
      'snackbar/showSnackbarSuccessAction': () => {
      },
      'snackbar/showSnackbarErrorAction': () => {
      },
    },
  });

  beforeEach(() => {
    wrapper = mount(DeviceRename, {
      store,
      localVue,
      stubs: ['fragment'],
      propsData: { name, uid },
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
    expect(wrapper.vm.name).toEqual(name);
    expect(wrapper.vm.uid).toEqual(uid);
  });
  it('Compare data with default value', () => {
    expect(wrapper.vm.dialog).toEqual(false);
    expect(wrapper.vm.invalid).toEqual(false);
    expect(wrapper.vm.editName).toEqual('39-5e-2a');
  });
  it('Show message tooltip to user owner', async (done) => {
    const icons = wrapper.findAll('.v-icon');
    const helpIcon = icons.at(0);
    helpIcon.trigger('mouseenter');
    await wrapper.vm.$nextTick();

    expect(icons.length).toBe(1);
    requestAnimationFrame(() => {
      expect(wrapper.find('[data-test="tooltipOwner-text"]').text()).toEqual('Edit');
      done();
    });
  });
});
