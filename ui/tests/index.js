import Vue from 'vue';
import Vuetify from 'vuetify';
import Clipboard from 'v-clipboard';
import {
  ValidationObserver,
  ValidationProvider,
} from 'vee-validate/dist/vee-validate.full';

import SnackbarError from '@/components/snackbar/SnackbarError';

Vue.component('SnackbarError', SnackbarError);
Vue.component('ValidationObserver', ValidationObserver);
Vue.component('ValidationProvider', ValidationProvider);

Vue.config.productionTip = false;
Vue.use(Vuetify);
Vue.use(Clipboard);
