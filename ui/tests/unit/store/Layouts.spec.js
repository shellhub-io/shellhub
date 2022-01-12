import store from '@/store';

describe('Stats', () => {
  const layouts = ['simpleLayout', 'appLayout'];

  it('Returns layout default variable', () => {
    expect(store.getters['layout/getLayout']).toEqual(layouts[1]);
  });
  it('Verify initial state change for setLayout mutation', () => {
    store.commit('layout/setLayout', layouts[0]);
    expect(store.getters['layout/getLayout']).toEqual(layouts[0]);
  });

  // Verify change dark to light mode

  it('verify light mode', () => {
    store.commit('layout/setStatusDarkMode', false);
    expect(store.getters['layout/getStatusDarkMode']).toEqual(false);
  });
});
