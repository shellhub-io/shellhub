import store from '@/store';

describe('Stats', () => {
  const stats = { registered_devices: 2, online_devices: 1, active_sessions: 1 };
  it('Returns stats default variable', () => {
    expect(store.getters['stats/stats']).toEqual({});
  });
  it('Verify initial state changes for mutation setStats', () => {
    store.commit('stats/setStats', { data: stats });
    expect(store.getters['stats/stats']).toEqual(stats);
  });
  it('Verify empty stats state for mutation clearListState', () => {
    store.commit('stats/clearListState', { data: stats });
    expect(store.getters['stats/stats']).toEqual({});
  });
});
