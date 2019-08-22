import { login } from '@/api/auth'

export default {
    namespaced: true,

    state: {
        status: '',
        token: localStorage.getItem('token') || '',
        user: {}
    },

    getters: {
        isLoggedIn: state => !!state.token,
        authStatus: state => state.status,
    },

    mutations: {
        auth_request(state) {
            state.status = 'loading'
        },

        auth_success(state, token, user){
            state.status = 'success'
            state.token = token
            state.user = user
        },

        auth_error(state){
            state.status = 'error'
        },

        logout(state){
            state.status = ''
            state.token = ''
        },
    },

    actions: {
        async login(context, user) {
            context.commit('auth_request')

            try {
                const resp = await login(user);

                localStorage.setItem('token', resp.data.token)
                context.commit('auth_success', resp.data.token)
            } catch (err) {
                alert(err);
            }
        },

        logout(context){
            context.commit('logout')
            localStorage.removeItem('token');
        }
    }
}