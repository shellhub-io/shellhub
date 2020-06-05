import store from '../../../src/store.js';

describe('Stats', () => {
  it('returns stats', () => {
    const actual = store.getters['stats/stats'];
    expect(actual).toEqual([]);
  });
  it('complete test', () => {
    const stats = {'registered_devices':2, 'online_devices':1, 'active_sessions':1};

    store.commit('stats/setStats', stats);

    expect(store.getters['stats/stats']).toEqual(stats);
  });
});
