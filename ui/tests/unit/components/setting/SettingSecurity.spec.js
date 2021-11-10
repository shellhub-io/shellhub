import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import SettingSecurity from '@/components/setting/SettingSecurity';
import { actions, authorizer } from '../../../../src/authorizer';

describe('SettingSecurity', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;

  const accessType = ['owner', 'administrator', 'operator', 'observer'];

  const hasAuthorization = {
    owner: true,
    administrator: true,
    operator: false,
    observer: false,
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
    },
  ];

  const storeVuex = (sessionRecord, currentAccessType) => new Vuex.Store({
    namespaced: true,
    state: {
      sessionRecord,
      currentAccessType,
    },
    getters: {
      'security/get': (state) => state.sessionRecord,
      'auth/accessType': (state) => state.currentAccessType,
    },
    actions: {
      'security/set': () => {},
      'security/get': () => {},
      'snackbar/showSnackbarErrorDefault': () => {},
    },
  });

  tests.forEach((test) => {
    accessType.forEach((currentAccessType) => {
      describe(`${test.description} ${currentAccessType}`, () => {
        beforeEach(() => {
          wrapper = shallowMount(SettingSecurity, {
            store: storeVuex(test.variables.sessionRecord, currentAccessType),
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
        it('Process data in the computed', () => {
          expect(wrapper.vm.hasAuthorization).toEqual(hasAuthorization[currentAccessType]);
        });
      });
    });
  });
});
