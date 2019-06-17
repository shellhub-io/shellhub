import Vue from 'vue'
import Vuex from 'vuex'

import devices from '@/modules/devices'
import modals from '@/modules/modals'
import stats from '@/modules/stats'
import sessions from '@/modules/sessions'

Vue.use(Vuex)

export default new Vuex.Store({
    modules: { devices, modals, stats, sessions }
})