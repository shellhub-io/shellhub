import Vue from 'vue';
import Vuex from 'vuex';
import stats from '@/modules/stats';
import sessions from '@/modules/sessions';
import auth from '@/modules/auth';
import devices from '@/modules/devices';
import modals from '@/modules/modals';
import firewallrules from '@/modules/firewall_rules';
import notifications from '@/modules/notifications';

Vue.use(Vuex);

export default new Vuex.Store({
  modules: {
    devices, modals, stats, sessions, auth, firewallrules, notifications,
  },
});
