import Vue from 'vue';
import Vuex from 'vuex';
import stats from '@/store/modules/stats';
import sessions from '@/store/modules/sessions';
import auth from '@/store/modules/auth';
import devices from '@/store/modules/devices';
import modals from '@/store/modules/modals';
import firewallrules from '@/store/modules/firewall_rules';
import notifications from '@/store/modules/notifications';

Vue.use(Vuex);

export default new Vuex.Store({
  modules: {
    devices, modals, stats, sessions, auth, firewallrules, notifications,
  },
});
