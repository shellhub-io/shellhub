import Vue from 'vue'
import { fetchSessions, getSession } from '@/api/sessions'

export default {
    namespaced: true,

    state: {
        sessions: [],
        session: []

    },

    getters: {
        list: state => state.sessions,
        get: state => state.session
    },

    mutations: {
        setSessions: (state, data) => {
            Vue.set(state, 'sessions', data)
        },
        setSession: (state, data) => {
            Vue.set(state, 'session', data)
        }
    },

    actions: {
        fetch: async (context) => {
            let res = await fetchSessions()

            context.commit('setSessions', res.data)
        },
        get: async (context,uid)  => {
            let res = await getSession(uid)
            context.commit('setSession', res.data)
        }

    }
}