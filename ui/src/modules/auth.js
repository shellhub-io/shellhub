import { login } from '@/api/auth'

export default {
    namespaced: true,

    state: {
        status: '',
        token: localStorage.getItem('token') || '',
        user: localStorage.getItem('user') || ''
    },

    getters: {
        isLoggedIn: state => !!state.token,
        authStatus: state => state.status,
        currentUser: state => state.user,
    },

    mutations: {
        auth_request(state) {
            state.status = 'loading'
        },

        auth_success(state, data) {
            state.status = 'success'
            state.token = data.token
            state.user = data.user
        },

        auth_error(state) {
            state.status = 'error'
        },

        logout(state) {
            state.status = ''
            state.token = ''
            state.user = ''
        },
    },

    actions: {
        async login(context, user) {
            context.commit('auth_request')

            try {
                const resp = await login(user);

                localStorage.setItem('token', resp.data.token)
                localStorage.setItem('user', resp.data.user)

                context.commit('auth_success', resp.data)
            } catch (err) {
                context.commit('auth_error')
            }
        },

        logout(context) {
            context.commit('logout')
            localStorage.removeItem('token');
            localStorage.removeItem('user');
        }
    }
}