export default {
    namespaced: true,

    state: {
        terminal: "",
        add_device: false
    },

    getters: {
        terminal: state => {
            return state.terminal
        },

        add_device: state => {
            return state.add_device
        }
    },

    mutations: {
        setTerminal: (state, data) => {
            state.terminal = data;
        },

        setAddDevice: (state, data) => {
            state.add_device = data;
        }
    },

    actions: {
        toggleTerminal: (context, value) => {
            context.commit('setTerminal', value)
        },

        showAddDevice: (context, value) => {
            context.commit('setAddDevice', value)
        }
    }
}