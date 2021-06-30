import store from '@/store';

describe('Snackbar', () => {
  const snackbarError = false;
  const snackbarSucess = false;
  const SnackbarMessageAndContentTypeError = { typeMessage: 'error', typeContent: '' };
  const SnackbarMessageAndContentTypeSuccess = { typeMessage: 'success', typeContent: '' };
  const SnackbarMessageAndContentType = { typeMessage: 'default', typeContent: '' };
  const SnackbarMessageAndContentTypeCopy = { typeMessage: '', typeContent: 'copy' };
  const SnackbarMessageAndContentTypeDefault = { typeMessage: '', typeContent: '' };
  const SnackbarMessageAndContentTypeAssociation = { typeMessage: 'association', typeContent: '' };
  const SnackbarMessageAndContentTypeIncorrect = { typeMessage: 'incorrect', typeContent: '' };
  const snackbarCopy = false;

  it('Return snackbar default variables', () => {
    expect(store.getters['snackbar/snackbarSuccess']).toEqual(snackbarSucess);
    expect(store.getters['snackbar/snackbarError']).toEqual(snackbarError);
    expect(store.getters['snackbar/snackbarMessageAndContentType']).toEqual(SnackbarMessageAndContentTypeDefault);
    expect(store.getters['snackbar/snackbarCopy']).toEqual(snackbarCopy);
  });
  it('Verify initial states change for mutation setSnackbarSuccessAction', () => {
    store.commit('snackbar/setSnackbarSuccessAction', SnackbarMessageAndContentTypeSuccess);
    expect(store.getters['snackbar/snackbarMessageAndContentType']).toEqual(SnackbarMessageAndContentTypeSuccess);
    expect(store.getters['snackbar/snackbarSuccess']).toEqual(true);
  });
  it('Verify initial states change for mutation showSnackbarSuccessDefault', () => {
    store.commit('snackbar/setSnackbarSuccessDefault');
    expect(store.getters['snackbar/snackbarMessageAndContentType']).toEqual(SnackbarMessageAndContentType);
    expect(store.getters['snackbar/snackbarSuccess']).toEqual(true);
  });
  it('Verify initial states change for mutation unsetSnackbarSuccess', () => {
    store.commit('snackbar/unsetSnackbarSuccess');
    expect(store.getters['snackbar/snackbarSuccess']).toEqual(false);
  });
  it('Verify initial states change for mutation setSnackbarErrorLoadingOrAction', () => {
    store.commit('snackbar/setSnackbarErrorLoadingOrAction', SnackbarMessageAndContentTypeError);
    expect(store.getters['snackbar/snackbarMessageAndContentType']).toEqual(SnackbarMessageAndContentTypeError);
    expect(store.getters['snackbar/snackbarError']).toEqual(true);
  });
  it('Verify initial states change for mutation setSnackbarErrorAssociation', () => {
    store.commit('snackbar/setSnackbarErrorAssociation', SnackbarMessageAndContentTypeAssociation);
    expect(store.getters['snackbar/snackbarMessageAndContentType']).toEqual(SnackbarMessageAndContentTypeAssociation);
    expect(store.getters['snackbar/snackbarError']).toEqual(true);
  });
  it('Verify initial states change for mutation setSnackbarErrorIncorrect', () => {
    store.commit('snackbar/setSnackbarErrorIncorrect', SnackbarMessageAndContentTypeIncorrect);
    expect(store.getters['snackbar/snackbarMessageAndContentType']).toEqual(SnackbarMessageAndContentTypeIncorrect);
    expect(store.getters['snackbar/snackbarError']).toEqual(true);
  });
  it('Verify initial states change for mutation setSnackbarErrorDefault', () => {
    store.commit('snackbar/setSnackbarErrorDefault');
    expect(store.getters['snackbar/snackbarMessageAndContentType']).toEqual(SnackbarMessageAndContentType);
    expect(store.getters['snackbar/snackbarError']).toEqual(true);
  });
  it('Verify initial states change for mutation unsetSnackbarError', () => {
    store.commit('snackbar/unsetSnackbarError');
    expect(store.getters['snackbar/snackbarError']).toEqual(false);
  });
  it('Verify initial states change for mutation setSnackbarCopy', () => {
    store.commit('snackbar/setSnackbarCopy', 'copy');
    expect(store.getters['snackbar/snackbarCopy']).toEqual(true);
    expect(store.getters['snackbar/snackbarMessageAndContentType']).toEqual(SnackbarMessageAndContentTypeCopy);
  });
  it('Verify initial states change for mutation unsetSnackbarCopy', () => {
    store.commit('snackbar/unsetSnackbarCopy');
    expect(store.getters['snackbar/snackbarCopy']).toEqual(false);
  });
});
