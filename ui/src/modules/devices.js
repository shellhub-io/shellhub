import Vue from 'vue'
import { fetchDevices, removeDevice, renameDevice, getDevice} from '@/api/devices'

export default {
    namespaced: true,

    state: {
        devices: [],
        device: []
    },

    getters: {
        list: state => state.devices,
        get: state => state.device
    },

    mutations: {
        setDevices: (state, data) => {
            Vue.set(state, 'devices', data)
        },

        removeDevice: (state, uid) => {
            state.devices.splice(state.devices.findIndex(d => d.uid == uid), 1)
        },

        renameDevice: (state, data) => {
            state.devices = state.devices.map(i => i.uid == data.uid ? { ...i, name: data.name } : i);
        },

        setDevice: (state, data) => {
            Vue.set(state, 'device', data)
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
        },

        rename: async (context, data) => {
            await renameDevice(data);
            context.commit('renameDevice', data)
        },

        get: async (context,uid)  => {
            let res = await getDevice(uid)
            context.commit('setDevice', res.data)
        }
    }
}