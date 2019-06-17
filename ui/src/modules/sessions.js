import Vue from 'vue'
import { fetchSessions } from '@/api/sessions'

export default {
    namespaced: true,

    state: {
        sessions: []
    },

    getters: {
        list: state => state.sessions
    },

    mutations: {
        setSessions: (state, data) => {
            Vue.set(state, 'sessions', data)
        }
    },

    actions: {
        fetch: async (context) => {
            let res = await fetchSessions()

            context.commit('setSessions', res.data)
        }
    }
}