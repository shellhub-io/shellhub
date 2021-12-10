import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import SettingSecurity from '@/components/setting/SettingSecurity';
import { actions, authorizer } from '../../../../src/authorizer';

describe('SettingSecurity', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;

  const role = ['owner', 'operator'];

  const hasAuthorization = {
    owner: true,
    operator: false,
  };

  const tests = [
    {
      description: 'SettingSecurity',
      variables: {
        sessionRecord: false,
      },
      props: {
        hasTenant: true,
      },
      data: {
        action: 'enableSessionRecord',
      },
    },
  ];

  const storeVuex = (sessionRecord, currentrole) => new Vuex.Store({
    namespaced: true,
    state: {
      sessionRecord,
      currentrole,
    },
    getters: {
      'security/get': (state) => state.sessionRecord,
      'auth/role': (state) => state.currentrole,
    },
    actions: {
      'security/set': () => {},
      'security/get': () => {},
      'snackbar/showSnackbarErrorDefault': () => {},
    },
  });

  tests.forEach((test) => {
    role.forEach((currentrole) => {
      describe(`${test.description} ${currentrole}`, () => {
        beforeEach(() => {
          wrapper = shallowMount(SettingSecurity, {
            store: storeVuex(test.variables.sessionRecord, currentrole),
            localVue,
            stubs: ['fragment'],
            propsData: { hasTenant: test.props.hasTenant },
            mocks: {
              $authorizer: authorizer,
              $actions: actions,
            },
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
        // Data checking
        //////

        it('Receive data in props', () => {
          Object.keys(test.props).forEach((item) => {
            expect(wrapper.vm[item]).toEqual(test.props[item]);
          });
        });
        it('Compare data with default value', () => {
          Object.keys(test.data).forEach((item) => {
            expect(wrapper.vm[item]).toEqual(test.data[item]);
          });
        });
        it('Process data in the computed', () => {
          expect(wrapper.vm.hasAuthorization).toEqual(hasAuthorization[currentrole]);
        });
      });
    });
  });
});
