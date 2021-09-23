import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import flushPromises from 'flush-promises';
import Vuetify from 'vuetify';
import DeviceActionButton from '@/components/device/DeviceActionButton';

describe('DeviceActionButton', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.use(Vuex);

  let wrapper;

  const isOwner = true;
  const isActive = true;
  const uid = '';
  let notificationStatus = true;
  let action = 'accept';

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      isOwner,
      isActive,
    },
    getters: {
      'namespaces/owner': (state) => state.isOwner,
      isActive: (state) => state.isActive,
    },
    actions: {
      'devices/refresh': () => {},
      'devices/accept': () => {},
      'devices/reject': () => {},
      'devices/remove': () => {},
      'notifications/fetch': () => {},
      'stats/get': () => {},
      'snackbar/showSnackbarErrorDefault': () => {},
      'snackbar/showSnackbarErrorAction': () => {},
      'snackbar/showSnackbarErrorLoading': () => {},
    },
  });

  beforeEach(() => {
    wrapper = mount(DeviceActionButton, {
      store,
      localVue,
      stubs: ['fragment'],
      propsData: { uid, notificationStatus, action },
      vuetify,
    });
  });

  ///////
  // Component Rendering
  //////

  it('Is a Vue instance', () => {
    document.body.setAttribute('data-app', true);
    expect(wrapper).toBeTruthy();
  });
  it('Renders the component', () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  ///////
  // Data and Props checking
  //////

  it('Receive data in props', () => {
    expect(wrapper.vm.uid).toEqual(uid);
    expect(wrapper.vm.notificationStatus).toEqual(notificationStatus);
    expect(wrapper.vm.action).toEqual(action);
  });
  it('Compare data with the default value', () => {
    expect(wrapper.vm.dialog).toEqual(false);
  });
  it('Process data in the computed', () => {
    expect(wrapper.vm.isOwner).toEqual(isOwner);
  });

  //////
  // HTML validation
  //////

  it('Renders the template with data', () => {
    expect(wrapper.find('[data-test="notification-btn"]').exists()).toBe(true);
  });

  it(`Renders the template with data - 
      notification is false,
      action is accept,
      dialog is false`,
  () => {
    notificationStatus = false;

    wrapper = mount(DeviceActionButton, {
      store,
      localVue,
      stubs: ['fragment'],
      propsData: { uid, notificationStatus, action },
      vuetify,
    });

    expect(wrapper.find('[data-test="notification-btn"]').exists()).toBe(false);
    expect(wrapper.find('[data-test="tooltipNotOwner-text"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="tooltipNotOwner-text"]').text()).toBe('accept');
  });
  it(`Renders the template with data - 
      notification is false,
      action is reject,
      dialog is false`,
  () => {
    notificationStatus = false;
    action = 'reject';

    wrapper = mount(DeviceActionButton, {
      store,
      localVue,
      stubs: ['fragment'],
      propsData: { uid, notificationStatus, action },
      vuetify,
    });

    expect(wrapper.find('[data-test="notification-btn"]').exists()).toBe(false);
    expect(wrapper.find('[data-test="tooltipNotOwner-text"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="tooltipNotOwner-text"]').text()).toBe('reject');
  });
  it(`Renders the template with data - 
      notification is false,
      action is remove,
      dialog is false`,
  () => {
    notificationStatus = false;
    action = 'remove';

    wrapper = mount(DeviceActionButton, {
      store,
      localVue,
      stubs: ['fragment'],
      propsData: { uid, notificationStatus, action },
      vuetify,
    });

    expect(wrapper.find('[data-test="notification-btn"]').exists()).toBe(false);
    expect(wrapper.find('[data-test="tooltipNotOwner-text"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="tooltipNotOwner-text"]').text()).toBe('remove');
  });
  it(`Renders the template with data -
    notification is false,
    action is accept,
    dialog is true`,
  async () => {
    notificationStatus = false;
    action = 'accept';

    wrapper = mount(DeviceActionButton, {
      store,
      localVue,
      stubs: ['fragment'],
      propsData: { uid, notificationStatus, action },
      vuetify,
    });

    wrapper.setData({ dialog: true });
    await flushPromises();

    const text = `You are about to ${action} this device.`;

    expect(wrapper.find('[data-test="deviceActionButton-card"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="dialog-text"]').text()).toBe(text);
    expect(wrapper.find('[data-test="dialog-text"]').text()).toBe(text);
    expect(wrapper.find('[data-test="cancel-btn"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="dialog-btn"]').text()).toBe(action);
  });
  it(`Renders the template with data -
    notification is false,
    action is accept,
    dialog is true`,
  async () => {
    notificationStatus = false;
    action = 'reject';

    wrapper = mount(DeviceActionButton, {
      store,
      localVue,
      stubs: ['fragment'],
      propsData: { uid, notificationStatus, action },
      vuetify,
    });

    wrapper.setData({ dialog: true });
    await flushPromises();

    const text = `You are about to ${action} this device.`;

    expect(wrapper.find('[data-test="deviceActionButton-card"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="dialog-text"]').text()).toBe(text);
    expect(wrapper.find('[data-test="dialog-text"]').text()).toBe(text);
    expect(wrapper.find('[data-test="cancel-btn"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="dialog-btn"]').text()).toBe(action);
  });
  it(`Renders the template with data -
    notification is false,
    action is accept,
    dialog is true`,
  async () => {
    notificationStatus = false;
    action = 'remove';

    wrapper = mount(DeviceActionButton, {
      store,
      localVue,
      stubs: ['fragment'],
      propsData: { uid, notificationStatus, action },
      vuetify,
    });

    wrapper.setData({ dialog: true });
    await flushPromises();

    const text = `You are about to ${action} this device.`;

    expect(wrapper.find('[data-test="deviceActionButton-card"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="dialog-text"]').text()).toBe(text);
    expect(wrapper.find('[data-test="dialog-text"]').text()).toBe(text);
    expect(wrapper.find('[data-test="cancel-btn"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="dialog-btn"]').text()).toBe(action);
  });
});
