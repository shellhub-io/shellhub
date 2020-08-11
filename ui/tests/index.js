import Vue from 'vue';
import Vuetify from 'vuetify';

import SnackbarError from '@/components/snackbar/SnackbarError';

Vue.component('SnackbarError', SnackbarError);

Vue.config.productionTip = false;
Vue.use(Vuetify);
