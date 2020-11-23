import Vue from 'vue';
import Vuex from 'vuex';
import stats from '@/store/modules/stats';
import sessions from '@/store/modules/sessions';
import auth from '@/store/modules/auth';
import devices from '@/store/modules/devices';
import modals from '@/store/modules/modals';
import snackbar from '@/store/modules/snackbar';
import firewallrules from '@/store/modules/firewall_rules';
import publickeys from '@/store/modules/public_keys';
import notifications from '@/store/modules/notifications';
import users from '@/store/modules/users';
import security from '@/store/modules/security';

Vue.use(Vuex);

export default new Vuex.Store({
  modules: {
    devices,
    modals,
    snackbar,
    stats,
    sessions,
    auth,
    firewallrules,
    publickeys,
    notifications,
    users,
    security,
  },
});
