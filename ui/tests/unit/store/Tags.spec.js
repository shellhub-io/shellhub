import store from '@/store';

describe('Tags', () => {
  const numberTags = 3;
  const tags = ['tag1', 'tag2', 'tag3'];

  it('Return public key default variables', () => {
    expect(store.getters['tags/list']).toEqual([]);
    expect(store.getters['tags/getNumberTags']).toEqual(0);
  });
  it('Verify initial state change for setTags mutation', () => {
    store.commit('tags/setTags', { data: tags, headers: { 'x-total-count': numberTags } });
    expect(store.getters['tags/list']).toEqual(tags);
    expect(store.getters['tags/getNumberTags']).toEqual(numberTags);
  });
});
