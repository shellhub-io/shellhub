import Vue from 'vue'

export default {
    namespaced: true,

    state: {
        terminal: ""
    },

    getters: {
        terminal: state => {
            return state.terminal
        }
    },

    mutations: {
        setTerminal: (state, data) => {
            state.terminal = data;
        }
    },

    actions: {
        toggleTerminal: (context, value) => {
            context.commit('setTerminal', value)
        }
    }
}