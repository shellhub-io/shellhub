import Vuex from 'vuex';
import { shallowMount, createLocalVue, config } from '@vue/test-utils';
import SettingNamespace from '@/components/setting/SettingNamespace';

config.mocks = {
  $env: {
    isEnterprise: true,
  },
};

describe('SettingNamespace', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;
  let wrapper2;
  let wrapper3;

  const namespace = {
    name: 'namespace3',
    members: [{ id: '6', name: 'user6' }, { id: '7', name: 'user7' }, { id: '8', name: 'user8' }],
    owner: '6',
    tenant_id: 'e359bf484715',
  };

  const idOwner = '6';
  const idNotOwner = '10';
  const textRole = ['Owner', 'Member', 'Member'];
  const owner = true;

  const storeNotOwner = new Vuex.Store({
    namespaced: true,
    state: {
      namespace,
      id: idNotOwner,
      owner: !owner,
    },
    getters: {
      'namespaces/get': (state) => state.namespace,
      'auth/id': (state) => state.id,
      'namespaces/owner': (state) => state.owner,
    },
    actions: {
      'namespaces/put': () => {
      },
      'namespaces/get': () => {
      },
      'namespaces/removeUser': () => {
      },
      'snackbar/showSnackbarSuccessAction': () => {
      },
      'snackbar/showSnackbarErrorAction': () => {
      },
    },
  });

  const storeOwner = new Vuex.Store({
    namespaced: true,
    state: {
      namespace,
      id: idOwner,
      owner,
    },
    getters: {
      'namespaces/get': (state) => state.namespace,
      'auth/id': (state) => state.id,
      'namespaces/owner': (state) => state.owner,
    },
    actions: {
      'namespaces/put': () => {
      },
      'namespaces/get': () => {
      },
      'namespaces/removeUser': () => {
      },
      'snackbar/showSnackbarSuccessAction': () => {
      },
      'snackbar/showSnackbarErrorAction': () => {
      },
    },
  });

  beforeEach(() => {
    jest.spyOn(Storage.prototype, 'getItem').mockReturnValue('e359bf484715');

    wrapper = shallowMount(SettingNamespace, {
      store: storeOwner,
      localVue,
      stubs: ['fragment'],
      mocks: ['$env'],
    });
    wrapper2 = shallowMount(SettingNamespace, {
      localVue,
      store: storeNotOwner,
      stubs: ['fragment'],
      mocks: ['$env'],
    });
    wrapper3 = shallowMount(SettingNamespace, {
      localVue,
      store: storeOwner,
      stubs: ['fragment'],
      mocks: {
        $env: {
          isEnterprise: false,
        },
      },
    });
  });

  it('Is a Vue instance', () => {
    expect(wrapper).toBeTruthy();
  });
  it('Renders the component', () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
  it('Renders the template with data', () => {
    expect(wrapper.find('[data-test=tenant]').text()).toEqual(namespace.tenant_id);
  });
  it('Process data in the computed', () => {
    expect(wrapper.vm.namespace).toEqual(namespace);
    expect(wrapper.vm.tenant).toEqual(namespace.tenant_id);
    expect(wrapper.vm.isEnterpriseOwner).toEqual(true);
  });
  it('Loads name when component is created', () => {
    wrapper.vm.$nextTick(() => {
      expect(wrapper.vm.name).toBe(namespace.name);
    });
  });
  it('Loads the owner in template', () => {
    expect(wrapper.find('[data-test=owner]').text()).toEqual('Owner');
  });
  namespace.members.forEach((member) => {
    it(`Loads ${member.name} member in template`, () => {
      expect(wrapper.find(`[data-test=${member.name}]`).text()).toEqual(member.name);
    });
  });
  // hosted version tests
  it('Check owner fields rendering in hosted version of the template', () => {
    expect(wrapper.find('[data-test=editOperation]').exists()).toEqual(true);
    expect(wrapper.find('[data-test=userOperation]').exists()).toEqual(true);
    expect(wrapper.find('[data-test=deleteOperation]').exists()).toEqual(true);
    expect(wrapper.find('[data-test=securityOperation]').exists()).toEqual(true);
    expect(wrapper.find('[data-test=notTheOwner]').exists()).toEqual(false);
    expect(wrapper.findAll('[data-test=remove-member]').length).toEqual(namespace.members.length - 1);
    expect(wrapper.find('[data-test=role]').exists()).toEqual(false);
    expect(wrapper.find('[data-test=new-member]').exists()).toEqual(true);
  });
  it('Check not the owner fields rendering in hosted version of the template.', () => {
    const notTheOwnerMessage = 'You\'re not the owner of this namespace.';
    expect(wrapper2.find('[data-test=editOperation]').exists()).toEqual(false);
    expect(wrapper2.find('[data-test=userOperation]').exists()).toEqual(false);
    expect(wrapper2.find('[data-test=deleteOperation]').exists()).toEqual(false);
    expect(wrapper2.find('[data-test=securityOperation]').exists()).toEqual(false);
    expect(wrapper2.find('[data-test=notTheOwner]').exists()).toEqual(true);
    expect(wrapper2.find('[data-test=notTheOwner]').text()).toEqual(notTheOwnerMessage);
  });
  // open version tests
  it('Check owner fields rendering in open version of the template.', () => {
    expect(wrapper3.find('[data-test=editOperation]').exists()).toEqual(true);
    expect(wrapper3.find('[data-test=userOperation]').exists()).toEqual(true);
    expect(wrapper3.find('[data-test=deleteOperation]').exists()).toEqual(true);
    expect(wrapper3.find('[data-test=securityOperation]').exists()).toEqual(false);
    expect(wrapper3.find('[data-test=notTheOwner]').exists()).toEqual(false);
    expect(wrapper3.findAll('[data-test=remove-member]').exists()).toEqual(false);
    expect(wrapper3.find('[data-test=role]').exists()).toEqual(true);
    expect(wrapper3.findAll('[data-test=role]').wrappers.reduce((ac, v) => {
      ac.push(v.text());
      return ac;
    }, [])).toEqual(textRole);
    expect(wrapper3.find('[data-test=new-member]').exists()).toEqual(false);
  });
});
