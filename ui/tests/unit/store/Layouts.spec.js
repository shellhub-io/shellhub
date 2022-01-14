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

  // Verify the default value of the navigation drawer status
  it('Returns navigation drawer default variable', () => {
    expect(store.getters['layout/getStatusNavigationDrawer']).toEqual(true);
  });
  it('Verify initial state change for setStatusNavigationDrawer mutation', () => {
    store.commit('layout/setStatusNavigationDrawer', false);
    expect(store.getters['layout/getStatusNavigationDrawer']).toEqual(false);
  });
});
