import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import DeviceIcon from '@/components/device/DeviceIcon';

describe('DeviceIcon', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;

  const iconName = 'alpine';
  const defaultIcon = 'fl-tux';

  const iconsMap = {
    alpine: 'fl-alpine',
    arch: 'fl-archlinux',
    centos: 'fl-centos',
    coreos: 'fl-coreos',
    debian: 'fl-debian',
    devuan: 'fl-devuan',
    elementary: 'fl-elementary',
    fedora: 'fl-fedora',
    freebsd: 'fl-freebsd',
    gentoo: 'fl-gentoo',
    linuxmint: 'fl-linuxmint',
    mageia: 'fl-mageia',
    manjaro: 'fl-manjaro',
    mandriva: 'fl-mandriva',
    nixos: 'fl-nixos',
    opensuse: 'fl-opensuse',
    rhel: 'fl-redhat',
    sabayon: 'fl-sabayon',
    slackware: 'fl-slackware',
    ubuntu: 'fl-ubuntu',
    raspbian: 'fl-raspberry-pi',
    'ubuntu-core': 'fl-ubuntu',
    void: 'fl-void',
  };

  beforeEach(() => {
    wrapper = shallowMount(DeviceIcon, {
      localVue,
      stubs: ['fragment'],
      propsData: { iconName },
    });
  });

  ///////
  // Component Rendering
  //////

  it('Is a Vue instance', () => {
    expect(wrapper).toBeTruthy();
  });
  it('Renders the component', () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  ///////
  // Data and Props checking
  //////

  it('Receive data in props', () => {
    expect(wrapper.vm.iconName).toBe(iconName);
  });
  it('Compare data with default value', () => {
    expect(wrapper.vm.deviceIcon).toEqual(iconsMap);
  });

  //////
  // HTML validation
  //////

  it('Renders the template with data', () => {
    //////
    // In this case, the default icon is tested.
    //////

    wrapper = shallowMount(DeviceIcon, {
      localVue,
      stubs: ['fragment'],
      propsData: { iconName: defaultIcon },
    });

    expect(wrapper.find('[data-test="type-icon"]').text()).toBe(defaultIcon);
  });

  //////
  // In this case, the other icons are tested.
  //////

  Object.keys(iconsMap).forEach((iconKey) => {
    wrapper = shallowMount(DeviceIcon, {
      localVue,
      stubs: ['fragment'],
      propsData: { iconName: iconKey },
    });

    expect(wrapper.find('[data-test="type-icon"]').text()).toBe(iconsMap[iconKey]);
  });
});
