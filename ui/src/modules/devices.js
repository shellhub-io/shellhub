import Vue from 'vue'
import { fetchDevices, removeDevice } from '@/api/devices'

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
        },

        removeDevice: (state, uid) => {
            state.devices.splice(state.devices.findIndex(d => d.uid == uid), 1)
        }
    },

    actions: {
        fetch: async (context) => {
            let res = await fetchDevices()

            context.commit('setDevices', res.data)
        },

        remove: async (context, uid) => {
            await removeDevice(uid);

            context.commit('removeDevice', uid)
        }
    }
}