import Vue from 'vue'
import { fetchDevices } from '@/api/devices'

export default {
    namespaced: true,

    state: {
        devices: []
    },

    getters: {
        list: state => state.devices
    },

    mutations: {
        setDevices: (state, data) => {
            Vue.set(state, 'devices', data)
        }
    },

    actions: {
        fetch: async (context) => {
            let res = await fetchDevices()

            context.commit('setDevices', res.data)
        }
    }
}