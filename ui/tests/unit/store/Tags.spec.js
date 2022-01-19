import store from '@/store';

describe('Tags', () => {
  const numberTags = 3;
  const tags = ['tag1', 'tag2', 'tag3'];
  const selected = ['tag1'];

  it('Return default variables', () => {
    expect(store.getters['tags/list']).toEqual([]);
    expect(store.getters['tags/getNumberTags']).toEqual(0);
    expect(store.getters['tags/selected']).toEqual([]);
  });
  it('Verify initial state change for setTags mutation', () => {
    store.commit('tags/setTags', { data: tags, headers: { 'x-total-count': numberTags } });
    expect(store.getters['tags/list']).toEqual(tags);
    expect(store.getters['tags/getNumberTags']).toEqual(numberTags);
  });

  it('Verify initial state change for setSelected mutation', () => {
    store.commit('tags/setSelected', selected);
    expect(store.getters['tags/selected']).toEqual(selected);
  });
  it('Verify clear change for setSelected mutation', () => {
    store.dispatch('tags/clearSelectedTags');
    expect(store.getters['tags/selected']).toEqual([]);
  });
});
